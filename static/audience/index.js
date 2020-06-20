const servers = {'iceServers': [{'urls': 'stun:stun.l.google.com:19302'}]};
const main_element = document.getElementsByTagName("main")[0];

const name = uuidv4();
const watchButton = document.getElementById("watch");
const watcher = document.getElementById("screen");

watchButton.onclick = () => {
    main_element.setAttribute("data-state", "playing");
	watcher.play();
}

function configureViewer(signaler, name) {
    const viewer = new RTCPeerConnection(servers);
    viewer.onicecandidate = ({candidate}) => signaler.send({candidate, from: name, to: "projectionist"});
    return viewer;
}

function gotRemoteStream(event) {
	main_element.setAttribute("data-state", "ready");
    if (watcher.srcObject !== event.streams[0]) {
        watcher.srcObject = event.streams[0];
    }
}

async function main(watcher) {
    const signaler = new Signal(name, "viewer");
	await signaler.configure();
    const viewer = configureViewer(signaler, name);
    viewer.ontrack = gotRemoteStream;
    signaler.onmessage = async ({ description, candidate, from, to }) => {
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

	await viewer.setLocalDescription(await viewer.createOffer());
	signaler.send({description: viewer.localDescription, from: name, to: "projectionist"});
}

main(watcher)

//https://stackoverflow.com/questions/105034/how-to-create-guid-uuid
function uuidv4() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}
