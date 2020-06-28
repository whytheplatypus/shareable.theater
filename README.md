# Sharable Theater

Watching things together while apart.



## Running locally

```
cd server; go run . -addr :<port number>
```

Example of compiling the scss files:

```
nvm use stable
npm install sass -g
cd static/projectionist
sass index.scss > index.css
```

## How it works

![Diagram](./drawing.svg)

**Less technical**

A projectionist creates a virtual screening room and makes a video file or part of their screen visible to those who join the room. The screening room's location is know by its URL which the projectionist has to share with viewers.

When a viewer joins by accessing the URL their (IP)address is shared to the projectionist, which the projectionist uses to stream to.

The projectionist has full control over the broadcast (play, pause, skip, etc.).

**More technical**

> This works by creating a p2p [RTCPeerConnection](https://developer.mozilla.org/en-US/docs/Web/API/RTCPeerConnection) between the host (or projectionist) and each member of the audience.
So the p2p aspect is a hub and spoke model. 
The video is captured either from a local video element if you're sharing a local file, or from some part of the projectionist's screen using `getDisplayMedia`, then streamed over that connection. 
There is a server that's responsible for the signaling required to create those p2p connections, but it's responsibilities end there.

~[whytheplatypus on HackerNews](https://news.ycombinator.com/item?id=23661310)


## known limitations

### Screen sharing audio

Sharing audio along with the screen will only work in chrome or chromium,
and then only when sharing other tabs with "share audio" enabled during selection.

### Video / audio encoding

Some encodings are better than others for the web.
Unfortunately I don't know enough to give reliable advice, so I recommend experimentation.

From some of local experiments:
- webm is great but vlc converted files tend to have strange problems with their time coding. (as far as I can tell this does not effect playback)
- mp4 works if the audio is encoded as mp3, I have noticed cases where the streamed aspect ratio is off.

### Bandwidth

The projectionist must have sufficient outward bandwidth to stream to multiple viewers.
Required bandwidth scales linearly with every viewer: `required_bandwidth = num_viewers * encoding_bitrate`.

As an example, take a [1080p 30fps video at 4500kbps](https://stream.twitch.tv/encoding/) and 3 viewers:
 `required_bandwith = 3 * 4500kbps = 13500kbps / 8 = 1687,5 KB/s / 1024 ~= 1.6 MB/s`

Additionaly WebRTC might use another bitrate and encoding that differs from what the file uses.
