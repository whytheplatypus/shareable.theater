const main_element = document.getElementsByTagName("main")[0];
const configuration = {'iceServers': [{'urls': 'stun:stun.l.google.com:19302'}]};
const video_source = document.getElementById("src");
const play_button = document.getElementById("play");
const player = document.getElementById("player");
video_source.addEventListener("input", function(ev) {
    main_element.setAttribute("data-state", "ready");
    document.getElementById("movie-name").innerHTML = ` ${video_source.files[0].name}`;
});
const loaded = new Promise((resolve, reject) => {
    play_button.addEventListener("click", function(ev) {
        ev.preventDefault();
        main_element.setAttribute("data-state", "playing");
        const video_source_url = window.URL.createObjectURL(video_source.files[0]);
    	player.src = video_source_url;
		player.addEventListener("loadeddata", function(){
        	resolve();
		});
    });
});

async function captureStream() {
	await loaded;
  	player.play();
    return player.captureStream();
}



function configure(host, signaler, peer) {
    host.onconnectionstatechange = e => console.debug(host.connectionState);
    host.onicecandidate = ({candidate}) => signaler.send({candidate, from: "host", to: peer});

    host.onnegotiationneeded = async () => {
        try {
            await host.setLocalDescription();
            signaler.send({ description: host.localDescription, from: "host", to: peer });
        } catch(err) {
            console.error(err);
        } 
    };

    host.onmessage = async ({ description, candidate, from, to }) => {
        let pc = host;

        try {
            if (description) {

                try {
                    await pc.setRemoteDescription(description);
                } catch(err) {
                    console.error(to, err);
                    return;
                } finally {
                    if (description.type =="offer") {
                        await pc.setLocalDescription();
                        signaler.send({description: pc.localDescription, from: to , to: from});
                    }
                }
            } else if (candidate != null) {
                await pc.addIceCandidate(candidate);
            }
        } catch(err) {
            console.error(err);
        }
    }

    return host;
}

let connections = {};

async function main() {
	console.debug("loading application");
    /*player.captureStream = player.captureStream || player.mozCaptureStream;*/
    /*const stream = player.captureStream()*/
    const stream = await captureStream();
    console.debug("got stream", stream);

    const signaler = new Signal("host");
	await signaler.configure();
	signaler.send({msg: "hello world"});

    signaler.onmessage = async (msg) => {
		console.debug("host got", msg);
        if (!(msg.from in connections)) {
            const host = new RTCPeerConnection(configuration);
            console.debug("created host");

            stream.getTracks().forEach(track => host.addTrack(track, stream));
            stream.addEventListener("addtrack", async (e) => {
                console.debug("track added");
                host.addTrack(e.track);
            });

            configure(host, signaler, msg.from);

            connections[msg.from] = host;
			return;
        }
        await connections[msg.from].onmessage(msg);
    };
}
window.onload = main;
