class Signal {
    constructor(name, role) {
		this.name = name;
    }

	async configure() {
		let self = this;
        this.conn = await connectAs(this.role); 
		this.conn.onmessage = function (evt) {
			var messages = evt.data.split('\n');
			for (var i = 0; i < messages.length; i++) {
				let msg = JSON.parse(messages[i])
				if (msg.to === self.name) {
					console.debug(self.name, "got", msg);
					self.onmessage(msg);
				}
			}
		};
	}

	onmessage(msg) {
		console.debug(msg);
	}

    send(msg) {
		this.conn.send(JSON.stringify(msg));
    }
}
