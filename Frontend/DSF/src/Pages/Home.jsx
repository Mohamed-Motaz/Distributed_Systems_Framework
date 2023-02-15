import React from "react";
import { UploadFileButtons } from "../components/UploadFileButtons.jsx";
import useWebSocket from "react-use-websocket";
import TextField from "@material-ui/core/TextField";
import { WebSocketServerService } from "./../services/WebSocketServerService";

export default function Home() {
  const WS_URL = "ws://localhost:3001/openWS/123";

  const wsClient = useWebSocket(WS_URL, {
    onOpen: () => {
      console.log("WebSocket connection established.");
    },
    onMessage: (e) => console.log({ e }),
  });

  return (
    <>
      <main>
        <UploadFileButtons wsClient={wsClient} />
      </main>
    </>
  );
}
