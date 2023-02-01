const { createContext, useState } = require("react");


export let AppContext = createContext(false);

export default function AppContextProvider(props) {
    const [isFlag, setIsFlag] = useState(false);

    function changeAppStauts(flag) {
        setIsFlag(flag);
    }

    return <AppContext.Provider value={{ isFlag, changeAppStauts }}>
        {props.children}
    </AppContext.Provider>
}