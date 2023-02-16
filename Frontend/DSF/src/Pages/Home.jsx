import React from "react";
import { UploadFileButtons } from "../components/UploadFileButtons.jsx";
import useWebSocket from "react-use-websocket";
import TextField from "@material-ui/core/TextField";
import { WebSocketServerService } from "./../services/WebSocketServerService";
import useAlert from "../helpers/useAlert.jsx";

export default function Home() {
  const WS_URL = "ws://localhost:3001/openWS/123";

  const wsClient = useWebSocket(WS_URL, {
    onOpen: () => {
      console.log("WebSocket connection established.");
    },
    onMessage: (e) => console.log({ e }),
  });

  const [AlertComponent, TriggerAlert] = useAlert();

  return (
    <>
      <main>
        <AlertComponent />

        <UploadFileButtons wsClient={wsClient} />

        <button className="m-16 p-4 rounded-lg bg-blue-800" onClick={() => TriggerAlert('My Alert Message')}>
          Trigger Alert
        </button>

      </main>
    </>
  );
}
