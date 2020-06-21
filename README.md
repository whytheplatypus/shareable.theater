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
