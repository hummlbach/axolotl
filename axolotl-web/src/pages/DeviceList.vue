<template>
  <div class="deviceList">
    <!-- The first device is the main device and should not be shown -->
    <div v-if="devices && devices.length > 1">
      <div v-for="(device, i) in devices" :key="device.id" class="row device">
        <div class="col-10">
          <div class="device-name">{{ device.name }}</div>
          <div class="meta">
            <span class="lastSeen">
              <span v-translate>Last seen:</span>
              {{ humanifyDate(device.lastSeen) }}
            </span>
          </div>
        </div>
        <div v-if="i!== 0" class="col-2 actions">
          <button class="btn" @click="delDevice(device.id)">
            <font-awesome-icon icon="trash" />
          </button>
        </div>
      </div>
    </div>
    <div v-else v-translate class="no-entries">No linked devices</div>
    <!-- eslint-enable -->

    <button class="btn start-chat" @click="linkDevice">
      <font-awesome-icon icon="plus" />
    </button>
    <add-device-modal
      v-if="showModal"
      @close="showModal = false"
      @add="addDevice($event)"
    />
  </div>
</template>

<script>
import AddDeviceModal from "@/components/AddDeviceModal";

export default {
  name: "DeviceList",
  components: {
    AddDeviceModal,
  },
  data() {
    return {
      showModal: false,
    };
  },
  computed: {
    devices() {
      return this.$store.state.devices;
    },
  },
  mounted() {
    this.$store.dispatch("getDevices");
  },
  methods: {
    linkDevice() {
      if (this.gui === "ut") {
        const result = window.prompt("desktopLink");
        this.showSettingsMenu = false;
        this.$store.dispatch("addDevice", result);
      } else {
        this.showModal = true;
      }
    },
    addDevice(qr) {
      this.showModal = false;
      if (qr !== "") this.$store.dispatch("addDevice", qr);
    },
    delDevice(id) {
      this.$store.dispatch("delDevice", id);
    },
    humanifyDate(inputDate) {
      const now = new Date();
      const date = new Date(inputDate);
      const diff = (now - date) / 1000;
      const seconds = diff;
      if (seconds < 60) return "now";
      const minutes = seconds / 60;
      if (minutes < 60) return Math.floor(minutes) + " minutes ago";
      const hours = minutes / 60;
      if (hours < 24) return Math.floor(hours) + " hours ago";
      return (
        date.getFullYear() +
        "-" +
        (date.getMonth() + 1) +
        "-" +
        date.getDate() +
        " " +
        date.getHours() +
        ":" +
        date.getMinutes() +
        ":" +
        date.getSeconds()
      );
    },
  },
};
</script>
<style scoped>
.device {
  border-bottom: 1px solid #c2c2c2;
  padding: 10px;
}
.actions {
  display: flex;
  justify-content: flex-end;
}
.lastSeen {
  font-size: 10px;
}
.device-name {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  max-width: 35ch;
}
</style>
