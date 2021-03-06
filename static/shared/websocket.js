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

function connect() {
	return new Promise((resolve, reject) => {
		const conn = new WebSocket(`ws://${window.location.host}${window.location.pathname}/signal`);
		conn.onopen = () => {
			resolve(conn);
		};
		conn.onerror = (e) => {
			console.debug(e);
		}
	})
}
