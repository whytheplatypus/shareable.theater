
let io = {};

class Signal {
    constructor(name) {
		this.name = name;
    }

	async configure() {
		let self = this;
        this.conn = await connectAs(this.name); 
		this.conn.onmessage = function (evt) {
			var messages = evt.data.split('\n');
			for (var i = 0; i < messages.length; i++) {
				let msg = JSON.parse(messages[i])
				console.debug(self.name, msg);
				self.onmessage(msg);
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
