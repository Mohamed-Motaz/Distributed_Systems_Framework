const { createContext, useState } = require("react");


export let AppContext = createContext(false);

export default function AppContextProvider(props) {
    const [apiEndPoint, setApiEndPoint] = useState('');



    return <AppContext.Provider value={{ apiEndPoint, setApiEndPoint }}>
        {props.children}
    </AppContext.Provider>
}