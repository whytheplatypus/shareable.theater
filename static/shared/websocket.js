const servers = {'iceServers': [
    {'urls': 'stun:stun.l.google.com:19302'},
    {'urls': 'stun:stun1.l.google.com:19302'},
    {'urls': 'stun:stunserver.org:19302'},
    {
        'urls': 'turn:numb.viagenie.ca',
        'credential': 'muazkh',
        'username': 'webrtc@live.com'
    },
]};

function connectAs(path) {
	return new Promise((resolve, reject) => {
		const conn = new WebSocket(`ws://${window.location.host}${window.location.pathname}/signal`);
		conn.onopen = () => {
			setInterval(function() {
				console.debug("pinging ws");
				conn.send("ping");
			}, 30000);
			resolve(conn);
		};
		conn.onclose = (e) => {
			console.debug("socket closed", e);
		};
		conn.onerror = (e) => {
			console.debug(e);
		}
	})
}
