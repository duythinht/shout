new Vue({
    el: "#app",
    data() {
      return {
        audio: null,
        circleLeft: null,
        barWidth: null,
        duration: null,
        currentTime: null,
        isTimerPlaying: false,
        tracks: [
          {
            name: "...",
            artist: "Radio Station",
            cover: "https://raw.githubusercontent.com/duythinht/shout/master/static/radio-girl.png",
            source: "https://radio.0x97a.com/stream.mp3",
            url: "https://github.com/duythinht/shout",
            favorited: false
          },
        ],
        currentTrack: null,
        currentTrackIndex: 0,
        transitionName: null
      };
    },
    methods: {
      play() {
        if (this.audio.paused) {
          this.audio.play();
          this.isTimerPlaying = true;
        } else {
          this.audio.pause();
          this.isTimerPlaying = false;
        }
      },
      generateTime() {
        this.barWidth = "98" + "%";
        this.circleLeft = "2"+ "%";
        let curmin = Math.floor(this.audio.currentTime / 60);
        let cursec = Math.floor(this.audio.currentTime - curmin * 60);
        if (curmin < 10) {
          curmin = "0" + curmin;
        }
        if (cursec < 10) {
          cursec = "0" + cursec;
        }
        this.duration = "∞";
        this.currentTime = curmin + ":" + cursec;
      },
      updateBar(x) {
        let progress = this.$refs.progress;
        let maxduration = this.audio.duration;
        let position = x - progress.offsetLeft;
        let percentage = (100 * position) / progress.offsetWidth;
        if (percentage > 100) {
          percentage = 100;
        }
        if (percentage < 0) {
          percentage = 0;
        }
        this.barWidth = percentage + "%";
        this.circleLeft = percentage + "%";
        this.audio.currentTime = (maxduration * percentage) / 100;
        this.audio.play();
      },
      clickProgress(e) {
        this.isTimerPlaying = true;
        this.audio.pause();
        this.updateBar(e.pageX);
      },
      prevTrack() {
        this.transitionName = "scale-in";
        this.isShowCover = false;
        if (this.currentTrackIndex > 0) {
          this.currentTrackIndex--;
        } else {
          this.currentTrackIndex = this.tracks.length - 1;
        }
        this.currentTrack = this.tracks[this.currentTrackIndex];
        this.resetPlayer();
      },
      nextTrack() {
        this.transitionName = "scale-out";
        this.isShowCover = false;
        if (this.currentTrackIndex < this.tracks.length - 1) {
          this.currentTrackIndex++;
        } else {
          this.currentTrackIndex = 0;
        }
        this.currentTrack = this.tracks[this.currentTrackIndex];
        this.resetPlayer();
      },
      resetPlayer() {
        this.barWidth = 0;
        this.circleLeft = 0;
        this.audio.currentTime = 0;
        this.audio.src = this.currentTrack.source;
        setTimeout(() => {
          if(this.isTimerPlaying) {
            this.audio.play();
          } else {
            this.audio.pause();
          }
        }, 300);
      },
      favorite() {
        this.tracks[this.currentTrackIndex].favorited = !this.tracks[
          this.currentTrackIndex
        ].favorited;
      }
    },
    created() {
      let vm = this;
      this.currentTrack = this.tracks[0];
      this.audio = new Audio();
      this.audio.src = this.currentTrack.source;
      this.audio.ontimeupdate = function() {
        vm.generateTime();
      };
      this.audio.onloadedmetadata = function() {
        vm.generateTime();
      };
      this.audio.onended = function() {
        vm.nextTrack();
        this.isTimerPlaying = true;
      };
      const ws = new WebSocket("wss://radio.0x97a.com/now-playing");
      ws.addEventListener("open", (event) => {
          console.log("Connected to server");
      });

      ws.addEventListener("close", (event) => {
          console.log("Connection closed", event.code, event.reason);
      });

      ws.addEventListener("error", (event) => {
          console.log("Connection error", event);
      });

      ws.addEventListener("message", (event) => {
          if (event.data.length <= 1) {
              return
          }

          try {
              const payload = JSON.parse(event.data);
              document.title = "Radio Station | " + payload.title;
              this.currentTrack.name = payload.title;
              setMediaSession(payload.title);
          } catch (e) {
              console.log("error: ", e)
              console.log(event.data.length)
              console.log(event.data)
          }
      });
    }
  });


  const setMediaSession = (title) => {
    if ("mediaSession" in navigator) {
        navigator.mediaSession.metadata = new MediaMetadata({
            title: title,
            artist: "Radio Station",
            album: "#music",
            artwork: [
                {
                    src: "https://raw.githubusercontent.com/duythinht/shout/master/static/96.png",
                    sizes: "96x96",
                    type: "image/png",
                },
                {
                    src: "https://raw.githubusercontent.com/duythinht/shout/master/static/128.png",
                    sizes: "128x128",
                    type: "image/png",
                },
                {
                    src: "https://raw.githubusercontent.com/duythinht/shout/master/static/192.png",
                    sizes: "192x192",
                    type: "image/png",
                },
                {
                    src: "https://raw.githubusercontent.com/duythinht/shout/master/static/256.png",
                    sizes: "256x256",
                    type: "image/png",
                },
                {
                    src: "https://raw.githubusercontent.com/duythinht/shout/master/static/384.png",
                    sizes: "384x384",
                    type: "image/png",
                },
                {
                    src: "https://raw.githubusercontent.com/duythinht/shout/master/static/512.png",
                    sizes: "512x512",
                    type: "image/png",
                },
            ],
        });
    }
};