class Signal {
    constructor(name, role) {
        this.name = name;
    }

    async configure() {
        let self = this;
        this.conn = await connect(); 
        this.conn.onclose = async (e) => {
            self.configure();
        };
        this.conn.onmessage = function (evt) {
            var messages = evt.data.split('\n');
            for (var i = 0; i < messages.length; i++) {
                try {
                    let msg = JSON.parse(messages[i])
                    if (msg.to === self.name) {
                        console.debug(self.name, "got", msg);
                        self.onmessage(msg);
                    }
                } catch (e) {
                    console.debug(e);
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
