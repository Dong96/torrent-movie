<template>
  <div>
    <CCard>
      <CCardHeader>
        <form @submit.prevent="addMagnet">
          <CRow>
            <CCol sm="1">
              <CButton color="primary" square type="submit">Add</CButton>
            </CCol>
            <CCol>
              <CInput placeholder="Magnet link ..." name="txtLink" />
            </CCol>
          </CRow>
        </form>
      </CCardHeader>
      <CCardBody>
        <CDataTable
          :items="items"
          :fields="fields"
          column-filter
          table-filter
          items-per-page-select
          :items-per-page="5"
          hover
          sorter
          pagination
        >
          <template #size="{item}">
            <td>{{getSize(item.size) }}</td>
          </template>

          <template #done="{item}">
            <td>
              <CProgress
                animated
                showPercentage
                showValue
                color="success"
                height="1.3rem"
                :precision="1"
                :value="getPercent(item)"
              />
            </td>
          </template>

          <template #encode="{item}">
            <td>
              <CProgress
                animated
                showPercentage
                showValue
                color="success"
                height="1.3rem"
                :precision="1"
                :value="item.encode.progress"
              />
            </td>
          </template>

          <template #status="{item}">
            <td>
              <CBadge :color="getBadge(item.state)">{{ getStatus(item) }}</CBadge>
            </td>
          </template>

          <template #dlspeed="{item}">
            <td>{{ getSpeed(item.dlspeed) }}</td>
          </template>

          <template #upspeed="{item}">
            <td>{{ getSpeed(item.upspeed) }}</td>
          </template>

          <template #estimate="{item}">
            <td>{{ getETA(item.eta) }}</td>
          </template>

          <template #downlstop="{item}">
            <td>
              <CButton
                v-if="item.progress != 1"
                class="mr-2 mb-1"
                color="primary"
                variant="outline"
                square
                size="sm"
                @click="downlStop(item, $event.target.textContent)"
              >{{ textDownButton(item) }}</CButton>
              <CButton
                class="mr-2 mb-1"
                color="primary"
                variant="outline"
                square
                size="sm"
                :disabled="isEncode(item)"
                @click="encodeVideo(item)"
              >{{ encodeButton(item) }}</CButton>
              <CButton
                class="mr-2 mb-1"
                color="primary"
                variant="outline"
                square
                size="sm"
                @click="deleteMovie(item)"
              >Delete</CButton>
              <CButton
                class="mr-2 mb-1"
                color="primary"
                variant="outline"
                square
                size="sm"
                :disabled="isEncode(item)"
                @click="watchMovie(item.name)"
              >Watch</CButton>
            </td>
          </template>
        </CDataTable>
      </CCardBody>
    </CCard>
  </div>
</template>

<script>
import api from "./api";
import utils from "@/utils/index";

const fields = [
  { key: "name", _style: "width: auto" },
  { key: "size", _style: "width: auto" },
  { key: "done", _style: "width: auto" },
  { key: "encode", _style: "width: auto" },
  { key: "status", _style: "width: auto" },
  { key: "dlspeed", label: "Down Speed", _style: "width: auto" },
  { key: "upspeed", label: "Up Speed", _style: "width: auto" },
  { key: "estimate", _style: "width: auto" },
  { key: "downlstop", label: "", _style: "width: auto" }
];

export default {
  name: "Dashboard",
  data() {
    return {
      items: null,
      fields
    };
  },
  mounted() {
    api.GetInitData().then(r => (this.items = r.data));
    api.InitWebSocket(data => {
      this.items = data.Torrents;
      // this.items.map(item => {
      //   let a = data.Torrents.find(t => t.hash == item.hash);
      //   console.log(a);
      //   return a;
      // });
    });
    // api.EncodeProgress(this.items);
  },
  methods: {
    watchMovie(name) {
      this.$router.push("/movie/" + name);
    },

    addMagnet: event =>
      api
        .AddMagnetLink(event.target.elements.txtLink.value)
        .then(r => console.log(r)),

    downlStop(item, state) {
      switch (state) {
        case "Pause":
          api.StopTorrent(item.hash);
          break;
        case "Resume":
          api.ResumeTorrent(item.hash);
          break;
      }
    },

    deleteMovie: item => api.DeleteTorrent(item.hash),

    getPercent(item) {
      const percent = item.progress * 100;
      return Math.round(percent * 10) / 10;
    },

    getStatus(item) {
      let s = "";
      switch (item.state) {
        case "downloading":
          s = "Downloading";
          break;
        case "pausedDL":
          s = "Paused";
          break;
        case "stalled":
          s = "Loading";
          break;
        case "stalledUP":
          s = "Done";
          break;
      }
      return s;
    },

    getBadge(status) {
      switch (status) {
        case "pausedDL":
          return "danger";
        case "stalled":
          return "warning";
        case "downloading":
          return "success";
        case "completed":
          return "success";
        case "stalledUP":
          return "success";
        default:
          "primary";
      }
    },

    isEncode: item => item.encode.progress == 1,

    encodeVideo: item => api.EncodeVideo(item.name),

    encodeButton: item => {
      if (item.encode.status == "continue") return "Stop";
      return "Encode";
    },

    getEncodeProgress: items =>
      api.getEncodeProgress(items, data => {
        items.encode = data;
      }),

    textDownButton: item => {
      switch (item.state) {
        case "downloading":
          return "Pause";
        case "pausedDL":
          return "Resume";
      }
      return "...";
    },

    getSize: size => utils.sizeConvert(size),

    getSpeed: speed => utils.speedConvert(speed),

    getETA: eta => utils.etaConvert(eta)
  },
  beforeDestroy: () => {
    api.CloseWB();
  }
};
</script>
