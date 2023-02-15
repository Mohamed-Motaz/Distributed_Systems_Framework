const { createContext, useState } = require("react");


export let AppContext = createContext(false);

export default function AppContextProvider(props) {
    const [isFlag, setIsFlag] = useState(false);
    const [apiEndPoint, setApiEndPoint] = useState('');

    function changeAppStauts(flag) {
        setIsFlag(flag);
    }

    return <AppContext.Provider value={{ setApiEndPoint }}>
        {props.children}
    </AppContext.Provider>
}