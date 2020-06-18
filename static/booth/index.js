const configuration = {'iceServers': [{'urls': 'stun:stun.l.google.com:19302'}]};

const main_element = document.getElementsByTagName("main")[0];
const video_source = document.getElementById("src");
const play_button = document.getElementById("play");
const player = document.getElementById("player");
const sharer = document.getElementById("sharer")
player.captureStream = player.captureStream || player.mozCaptureStream;
function playVideo(file) {
	main_element.setAttribute("data-state", "playing");

	const video_source_url = URL.createObjectURL(file);
    player.src = video_source_url;
}
play_button.addEventListener("click", function(ev) {
	ev.preventDefault();
    playVideo(video_source.files[0]);
});

// there is a bug in firefox
// the mozCaptureStream() call removes the audio from the video element and stream.
// Likely this will be fixed if and when mozCaptureStream stops being prefixed
// For now using a proxi video element for local playback seems to solve this
sharer.srcObject = player.captureStream()
const stream = player.captureStream()
stream.onaddtrack = updateTracks;


video_source.addEventListener("change", function(ev) {
    main_element.setAttribute("data-state", "ready");
    document.getElementById("movie-name").innerHTML = ` ${video_source.files[0].name}`;
});


function configure(host, signaler, peer) {
    host.onconnectionstatechange = e => console.debug(host.connectionState);
    host.onicecandidate = ({candidate}) => signaler.send({candidate, from: "host", to: peer});

    host.onnegotiationneeded = async () => {
        try {
            await host.setLocalDescription(await host.createOffer());
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
                        await pc.setLocalDescription(await pc.createAnswer());
                        signaler.send({description: pc.localDescription, from: to , to: from});
                    }
                }
            } else if (candidate) {
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
	for (let conn in connections) {
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
            connections[msg.from] = host;
            console.debug("created host");

            stream.getTracks().forEach(track => host.addTrack(track));

            configure(host, signaler, msg.from);

			return;
        }
        await connections[msg.from].onmessage(msg);
    };
}
main();

const audienceLink = document.querySelector("#audience-link");
const audienceUrl = window.location.href.replace("booth", "audience");
audienceLink.value = audienceUrl;

function copy() {
  audienceLink.select();
  document.execCommand("copy");
}

document.querySelector("#copy").onclick = copy;
