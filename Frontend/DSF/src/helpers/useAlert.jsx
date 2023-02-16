import React, { useEffect, useState } from 'react'

export default function useAlert() {

    const [alert, setAlert] = useState('')
    const [timer, setTimer] = useState(null)

    function TriggerAlert(message){
        clearTimeout(timer)
        setAlert(message)
        setTimer(setTimeout(() => setAlert(''), 2000))
    }

    const AlertComponent = () => <div className={`fixed bottom-8 right-8 z-50 max-w-sm transition-all duration-500 ${alert ? 'opacity-100' : 'opacity-0'} `}>
        <div className={`bg-black flex items-center justify-center w-fit rounded-3xl`}>
            <div className="px-4 py-5 text-sm text-white rounded-3xl bg-red-800 ">
                ALERT
            </div>
            <div className={`${alert ? 'pl-3 pr-5' : 'opacity-100'} line-clamp-3 text-sm`}>
                {alert}
            </div>
        </div>
    </div>

    return [AlertComponent, TriggerAlert]

}

