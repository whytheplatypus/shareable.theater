const main_element = document.getElementsByTagName("main")[0];
const configuration = {'iceServers': [{'urls': 'stun:stun.l.google.com:19302'}]};

const name = uuidv4();
const watchButton = document.getElementById("watch");
const watcher = document.getElementById("screen");
watchButton.onclick = () => {
    main_element.setAttribute("data-state", "playing");
    watcher.play();
}

function configureViewer(signaler, name) {
    const viewer = new RTCPeerConnection(configuration);
    viewer.onicecandidate = ({candidate}) => signaler.send({candidate, from: name, to: "host"});
    return viewer;
}

//https://stackoverflow.com/questions/105034/how-to-create-guid-uuid
function uuidv4() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

let inboundStream;

async function addWatcher(watcher) {
    const signaler = new Signal("viewer");
	await signaler.configure();
    const viewer = configureViewer(signaler, name);
    function gotRemoteStream(event) {
        main_element.setAttribute("data-state", "ready");
		console.debug(event.track);
		if (!inboundStream) {
			inboundStream = new MediaStream();
			watcher.srcObject = inboundStream;
		}
		// hack: assumes 1 movie = 2 tracks
		if (inboundStream.getTracks().length > 1) {
			inboundStream.getTracks().forEach(track => inboundStream.removeTrack(track));
		}
		inboundStream.addTrack(event.track);

		console.debug(inboundStream.getTracks());
    }
    viewer.ontrack = gotRemoteStream;
    signaler.onmessage = async ({ description, candidate, from, to }) => {
		if (to != name) {
			return;
		}
        let pc = viewer;
        try {
            if (description) {

                try {
                    await pc.setRemoteDescription(description);
                } catch(err) {
                    await Promise.all([
                        pc.setLocalDescription({type: "rollback"}),
                        pc.setRemoteDescription(description)
                    ]);
                } finally {
                    if (description.type == "offer") {
                        console.debug(description);
                        console.debug(to, "accepting offer");
                        await pc.setLocalDescription();
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
	await viewer.setLocalDescription();
	signaler.send({description: viewer.localDescription, from: name, to: "host"});
}

addWatcher(watcher)
