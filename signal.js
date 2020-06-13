
let io = {};

class Signal {
    constructor(name) {
        this.name = name; 
        io[name] = this;
    }

    send(msg) {
        if (msg.to in io) {
            io[msg.to].recievedMessage(msg);
        }
    }

    recievedMessage(msg) {
        console.debug(msg);
        this.onmessage(msg);
    }
}
