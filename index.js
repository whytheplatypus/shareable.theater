const offerOptions = {
    offerToReceiveAudio: 1,
    offerToReceiveVideo: 1
};
const video_source = document.getElementById("src");
const play_button = document.getElementById("play");
const player = document.getElementById("player");
const loaded = new Promise((resolve, reject) => {
    play_button.addEventListener("click", function(ev) {
        ev.preventDefault();
        const video_source_url = window.URL.createObjectURL(video_source.files[0]);
        resolve(video_source_url);
    });
});

async function captureStream() {
    player.src = await loaded;
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
    player.captureStream = player.captureStream || player.mozCaptureStream;
    const stream = player.captureStream()
    console.debug("got stream", stream);
    captureStream();

    const signaler = new Signal("host");
	await signaler.configure();
	signaler.send({msg: "hello world"});

    signaler.onmessage = async (msg) => {
		console.debug("host got", msg);
        if (!(msg.from in connections)) {
            const host = new RTCPeerConnection(null);
            console.debug("created host");

            stream.getTracks().forEach(track => host.addTrack(track, stream));
            stream.addEventListener("addtrack", async (e) => {
                console.debug("track added");
                host.addTrack(e.track, stream);
            });

            configure(host, signaler, msg.from);

            connections[msg.from] = host;
			return;
        }
        await connections[msg.from].onmessage(msg);
    };
}
window.onload = main;
