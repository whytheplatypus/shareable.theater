const video_form = document.getElementById("videos");
const video_source = document.getElementById("src");
const player = document.getElementById("player");
video_form.addEventListener("submit", function(ev) {
    ev.preventDefault();
    const video_source_url = window.URL.createObjectURL(video_source.files[0]);
    player.src = video_source_url;
    player.play();
});
