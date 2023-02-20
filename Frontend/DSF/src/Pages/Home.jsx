import React, { useContext } from "react";
import { UploadFileButtons } from "../components/UploadFileButtons.jsx";
import useWebSocket from "react-use-websocket";
import TextField from "@material-ui/core/TextField";
import { WebSocketServerService } from "./../services/WebSocketServerService";
import useAlert from "../helpers/useAlert.jsx";
import { Dropdown } from "flowbite-react";
import { AppContext } from "../context/AppContext.js";

export default function Home() {
  const { clientId, apiEndPoint } = useContext(AppContext);
  const WS_URL = `ws://${apiEndPoint}/openWS/${clientId}`;

  const wsClient = useWebSocket(WS_URL, {
    onOpen: () => {
      console.log("WebSocket connection established.");
    },
    onClose: () => {
      console.log('WebSocket connection closed, it will be re-established in a second');
      setTimeout(() => wsClient(), 1000);
    },
    onMessage: (e) => console.log({ e }),
  });

  const [AlertComponent, TriggerAlert] = useAlert();

  return (
    <>
      <main>
        <AlertComponent success={false} />

        <UploadFileButtons wsClient={wsClient} />

        <button className="m-16 p-4 rounded-lg bg-blue-800" onClick={() => TriggerAlert('My Alert Message')}>
          Trigger Alert
        </button>

      </main>
    </>
  );
}
