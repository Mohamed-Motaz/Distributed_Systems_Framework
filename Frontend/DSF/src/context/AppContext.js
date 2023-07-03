// import { uuid } from "react-uuid";
import { WebSocketServerService } from "../services/WebSocketServerService.js";

const { createContext, useState } = require("react");

export let AppContext = createContext(false);

export default function AppContextProvider(props) {
  const storedApiEndPoint = localStorage.getItem("apiEndPoint");

  const [apiEndPoint, setApiEndPoint] = useState(storedApiEndPoint || "");

  async function changeApiEndPoint(endPoint) {
    if (endPoint.includes("://")) {
      endPoint = endPoint.split("://")[1];
    }

    localStorage.setItem("apiEndPoint", endPoint);
    const isAlive = await WebSocketServerService().pingEndPoint();

    if (!isAlive) {
      localStorage.removeItem("apiEndPoint");
      return false;
    }

    setApiEndPoint(endPoint);

    return isAlive;
  }

  if (!sessionStorage.getItem("clientId")) {
    sessionStorage.setItem(
      "clientId",
      JSON.stringify(
        Date.now().toString(36) + Math.random().toString(36).substr(2)
      )
    );
  }

  return (
    <AppContext.Provider
      value={{
        apiEndPoint,
        changeApiEndPoint,
        clientId: JSON.parse(sessionStorage.getItem("clientId")),
      }}
    >
      {props.children}
    </AppContext.Provider>
  );
}
