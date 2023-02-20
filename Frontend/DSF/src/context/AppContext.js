// import { uuid } from "react-uuid";

const { createContext, useState } = require("react");

export let AppContext = createContext(false);

export default function AppContextProvider(props) {
  const [apiEndPoint, setApiEndPoint] = useState("");
  const [clientId, setClientId] = useState(
    Date.now().toString(36) + Math.random().toString(36).substr(2)
  );

  return (
    <AppContext.Provider
      value={{ apiEndPoint, setApiEndPoint, clientId, setClientId }}
    >
      {props.children}
    </AppContext.Provider>
  );
}
