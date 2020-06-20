function connectAs(path) {
	return new Promise((resolve, reject) => {
		const conn = new WebSocket(`wss://${window.location.host}${window.location.pathname}/signal`);
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
