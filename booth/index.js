const configuration = {'iceServers': [{'urls': 'stun:stun.l.google.com:19302'}]};

const main_element = document.getElementsByTagName("main")[0];
const video_source = document.getElementById("src");
const play_button = document.getElementById("play");
const player = document.getElementById("player");
player.captureStream = player.captureStream || player.mozCaptureStream;
var stream = player.captureStream()
stream.onaddtrack = updateTracks;

video_source.addEventListener("input", function(ev) {
    main_element.setAttribute("data-state", "ready");
    document.getElementById("movie-name").innerHTML = ` ${video_source.files[0].name}`;
});

play_button.addEventListener("click", function(ev) {
	ev.preventDefault();
	
	main_element.setAttribute("data-state", "playing");

	const video_source_url = window.URL.createObjectURL(video_source.files[0]);
	player.src = video_source_url;

	stream = player.captureStream()
	stream.onaddtrack = updateTracks;

  	player.play();
});

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

function updateTracks(e) {
	for (conn in connections) {
    	connections[conn].addTrack(e.track);
	}
}

async function main() {
	console.debug("loading application");

    const signaler = new Signal("host");
	await signaler.configure();
	signaler.send({msg: "hello world"});

    signaler.onmessage = async (msg) => {
		console.debug("host got", msg);
        if (!(msg.from in connections)) {
            const host = new RTCPeerConnection(configuration);
            console.debug("created host");

            stream.getTracks().forEach(track => host.addTrack(track));

            configure(host, signaler, msg.from);

            connections[msg.from] = host;
			return;
        }
        await connections[msg.from].onmessage(msg);
    };
}
main();
