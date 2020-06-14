const addWatcherButton = document.getElementById("add-watcher");
addWatcherButton.addEventListener("click", function(e) {
	e.preventDefault();
    let watcher = document.createElement("video"); 
    document.body.appendChild(watcher);
    addWatcher(watcher)
});

/*setTimeout(function() {
    let watcher = document.createElement("video"); 
    document.body.appendChild(watcher);
    addWatcher(watcher)
}, 5000);*/

function configureViewer(signaler, name) {
    const viewer = new RTCPeerConnection(null);
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

async function addWatcher(watcher) {
    const name = uuidv4();
    const signaler = new Signal("viewer");
	await signaler.configure();
    const viewer = configureViewer(signaler, name);
    function gotRemoteStream(event) {
        if (watcher.srcObject !== event.streams[0]) {
            console.debug("got remote stream", event.streams[0].getTracks());
            watcher.srcObject = event.streams[0];
            watcher.play();
        }
    }
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
