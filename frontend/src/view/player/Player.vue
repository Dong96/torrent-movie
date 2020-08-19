<template>
  <CRow alignHorizontal="center">
    <div id="videoPlayer"></div>
  </CRow>
</template>

<script>
export default {
  name: "Player",
  props: ["movie"],
  mounted() {
    const url = `http://localhost:8081/movie/${this.movie}/dash/manifest.mpd`;
    // eslint-disable-next-line no-undef
    jwplayer("videoPlayer").setup({
      playlist: [
        {
          file: url
        }
      ],
      width: 960,
      height: 590
    });
  },
  data() {
    return {};
  }
};
</script>