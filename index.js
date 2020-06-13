const offerOptions = {
  offerToReceiveAudio: 1,
  offerToReceiveVideo: 1
};
const video_form = document.getElementById("videos");
const video_source = document.getElementById("src");
const player = document.getElementById("player");
const watcher = document.getElementById("watcher");
const loaded = new Promise((resolve, reject) => {
	video_form.addEventListener("submit", function(ev) {
		ev.preventDefault();
		const video_source_url = window.URL.createObjectURL(video_source.files[0]);
		resolve(video_source_url);
	});
});
async function main() {
	player.src = await loaded;
	console.debug(player);

	player.captureStream = player.captureStream || player.mozCaptureStream;
	const stream = player.captureStream();
	console.debug("capture stream", stream);
	const host = new RTCPeerConnection(null);
	console.debug("created host");
	host.onconnectionstatechange = e => console.debug(e);

	const viewer = new RTCPeerConnection(null);
	function gotRemoteStream(event) {
  		if (watcher.srcObject !== event.streams[0]) {
			console.debug("got remote stream", event.streams);
			watcher.srcObject = event.streams[0];
  		}
	}
	viewer.ontrack = gotRemoteStream;

	viewer.onicecandidate = ({candidate}) => host.addIceCandidate(candidate);
	host.onicecandidate = ({candidate}) => viewer.addIceCandidate(candidate);

	host.onnegotiationneeded = async () => {
		console.debug("test");
	};
	stream.getTracks().forEach(track => console.debug(track));
	console.debug("loaded hosts movie");

	let not_connected = true;
	stream.onaddtrack = async (e) => {
		console.debug("track added");
		host.addTrack(e.track, stream);
		if (not_connected) {
			not_connected = false;
			console.debug("negotiation needed start");
			await host.setLocalDescription();
			console.debug("host formulates an answer to viewers call");
			await viewer.setRemoteDescription(host.localDescription);
			console.debug("viewer sees that host has picked up, sets remote description");
			await viewer.setLocalDescription();
			console.debug("viewer starting to call into host");
			await host.setRemoteDescription(viewer.localDescription)
			console.debug("host gets viewers call, sets remote description");
		}
	};
}
main();
