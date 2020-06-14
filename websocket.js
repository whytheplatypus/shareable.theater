function connectAs(path) {
	return new Promise((resolve, reject) => {
		const conn = new WebSocket(`ws://${window.location.host}/${path}`);
		conn.onopen = () => {
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
