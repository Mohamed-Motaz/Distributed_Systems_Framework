// import { uuid } from "react-uuid";
import { WebSocketServerService } from "../services/WebSocketServerService.js";

const { createContext, useState } = require("react");

export let AppContext = createContext(false);

export default function AppContextProvider(props) {

  const storedApiEndPoint = localStorage.getItem('apiEndPoint')

  const [apiEndPoint, setApiEndPoint] = useState(storedApiEndPoint || "")

  async function changeApiEndPoint(endPoint) {

    const isAlive = await WebSocketServerService().pingEndPoint(endPoint)
    if (!isAlive) { return isAlive }
    
    endPoint = endPoint.split('://')[1]
    setApiEndPoint(endPoint)
    localStorage.setItem('apiEndPoint', endPoint)
    return isAlive
  }

  const [clientId, setClientId] = useState(
    Date.now().toString(36) + Math.random().toString(36).substr(2)
  );

  return (
    <AppContext.Provider
      value={{ apiEndPoint, changeApiEndPoint, clientId, setClientId }}
    >
      {props.children}
    </AppContext.Provider>
  );
}
