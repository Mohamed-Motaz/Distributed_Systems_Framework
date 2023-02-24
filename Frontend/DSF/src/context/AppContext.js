// import { uuid } from "react-uuid";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import { handleGetAllBinaries } from "../services/ServiceTypes/HandlerGroup.js";
import useAlert from "../helpers/useAlert.jsx";

const { createContext, useState } = require("react");

export let AppContext = createContext(false);

export default function AppContextProvider(props) {
  const storedApiEndPoint = localStorage.getItem("apiEndPoint");

  const [apiEndPoint, setApiEndPoint] = useState(storedApiEndPoint || "");
  const [binaries, setBinaries] = useState({
    process: [],
    aggregate: [],
    distribute: [],
  });
  const [AlertComponent, TriggerAlert] = useAlert();
  const [isSuccess, setIsSuccess] = useState(false);

  async function changeApiEndPoint(endPoint) {
    if (endPoint.includes("://")) {
      endPoint = endPoint.split("://")[1];
    }
    setApiEndPoint(endPoint);
    localStorage.setItem("apiEndPoint", endPoint);

    const isAlive = await WebSocketServerService().pingEndPoint();

    if (!isAlive) {
      localStorage.removeItem("apiEndPoint")
      return false;
    }

    return isAlive;
  }

  const setAllBinaries = async () => {
    const files = await handleGetAllBinaries(TriggerAlert, setIsSuccess);

    const { AggregateBinaryNames, ProcessBinaryNames, DistributeBinaryNames } =
      files?.data?.response;
    setBinaries({
      process: ProcessBinaryNames,
      aggregate: AggregateBinaryNames,
      distribute: DistributeBinaryNames,
    });
  };

  const [clientId, setClientId] = useState(
    Date.now().toString(36) + Math.random().toString(36).substr(2)
  );

  return (
    <AppContext.Provider
      value={{
        apiEndPoint,
        changeApiEndPoint,
        clientId,
        setAllBinaries,
        binaries,
        setClientId,
        isSuccess,
        setIsSuccess,
        AlertComponent,
        TriggerAlert,
      }}
    >
      {props.children}
    </AppContext.Provider>
  );
}
