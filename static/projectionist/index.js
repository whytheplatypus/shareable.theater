const main_element = document.getElementsByTagName("main")[0];
const video_source_input = document.getElementById("src");
const play_button = document.getElementById("play");
const player = document.getElementById("player");
const share_screen = document.getElementById("share-screen");
// there is a bug in firefox
// the mozCaptureStream() call removes the audio from the video element and stream.
// Likely this will be fixed if and when mozCaptureStream stops being prefixed
// For now using a proxi video element for local playback seems to solve this
player.captureStream = player.captureStream || function() {
    const fallback_player = document.getElementById("fallback-player")
    const stream = player.mozCaptureStream.apply(arguments);
    fallback_player.srcObject = stream;
    return stream;
}.bind(player);
let remoteStream;
let controlStream;

share_screen.addEventListener("click", async function(ev) {
    const gdmOptions = {
        video: true,
        audio: true
    };
    let captureStream = null;

    try {
        captureStream = await navigator.mediaDevices.getDisplayMedia(gdmOptions);
    } catch(err) {
        console.error("Error: " + err);
        return
    }
    let tracks = captureStream.getTracks()
    for (let track in tracks) {
        tracks[track].addEventListener('ended', e => {
            console.debug(e);
            console.debug('Capture stream inactive - stop streaming!');
            main_element.setAttribute("data-state", "initial");
        });
    }
    remoteStream = new MediaStream();
    player.srcObject = captureStream;
    controlStream = captureStream;
    controlStream.onaddtrack = updateTracks;
    player.load();
});

play_button.addEventListener("click", function(ev) {
    ev.preventDefault();
    const video_source_url = URL.createObjectURL(video_source_input.files[0]);
    player.src = video_source_url;
    remoteStream = new MediaStream();
    controlStream = player.captureStream()
    controlStream.onaddtrack = updateTracks;
    player.load();
});

// states
video_source_input.addEventListener("change", function(ev) {
    main_element.setAttribute("data-state", "ready");
    document.getElementById("movie-name").innerHTML = ` ${video_source_input.files[0].name}`;
});

player.addEventListener("play", function() {
    main_element.setAttribute("data-state", "playing");
})


const connections = {};

function updateTracks(e) {
    for (let conn in connections) {
        connections[conn].addTrack(e.track, remoteStream);
    }
}


async function main() {
    console.debug("loading application");
    const signaler = new Signal("projectionist", "projectionist");
    await signaler.configure();

    signaler.onmessage = async (msg) => {
        if (!(msg.from in connections)) {
            const projectionist = new RTCPeerConnection(servers);

            connections[msg.from] = projectionist;

            projectionist.onconnectionstatechange = function(event) {
                console.debug(msg.from, "connection state:", event);
                switch(projectionist.connectionState) {
                    case "disconnected":
                    case "failed":
                    case "closed":
                        delete connections[msg.from];
                        break;
                }
            }

            if (controlStream) {
                controlStream.getTracks().forEach(track => projectionist.addTrack(track, remoteStream));
            }

            configure(projectionist, signaler, msg.from);

            return;
        }
        await connections[msg.from].onmessage(msg);
    };

}
main();

function configure(projectionist, signaler, peer) {
    projectionist.onicecandidate = ({candidate}) => signaler.send({candidate, from: "projectionist", to: peer});

    projectionist.onnegotiationneeded = async () => {
        try {
            await projectionist.setLocalDescription(await projectionist.createOffer());
            signaler.send({ description: projectionist.localDescription, from: "projectionist", to: peer });
        } catch(err) {
            console.error(err);
        } 
    };

    projectionist.onmessage = async ({ description, candidate, from, to }) => {
        let pc = projectionist;

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

    return projectionist;
}
