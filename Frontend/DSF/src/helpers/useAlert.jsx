import React, { useState } from "react";

export default function useAlert() {
  const [alert, setAlert] = useState("");
  const [timer, setTimer] = useState(null);
  const [handleOnClick, setHandleOnClick] = useState(null);

  function TriggerAlert(message, onClick = null) {
    clearTimeout(timer);
    setAlert(message);
    setHandleOnClick(onClick);
    setTimer(
      setTimeout(() => {
        setAlert("");
        setHandleOnClick(null);
      }, 5000)
    );
  }

  const AlertComponent = ({ success }) => (
    <div
      className={`fixed bottom-8 right-8 z-50 max-w-sm transition-all duration-500 ${
        alert ? "opacity-100" : "opacity-0"
      } `}
    >
      <div
        className={`bg-black flex items-center justify-center w-fit rounded-3xl`}
      >
        <div
          className={`px-4 py-5 text-sm text-white rounded-3xl ${
            handleOnClick ? "w-2/4" : "w-3/4"
          } ${success ? "bg-green-800" : "bg-red-800"} `}
        >
          ALERT
        </div>
        <div
          className={`${
            alert ? "pl-3 pr-5" : "opacity-100"
          } line-clamp-3 text-sm`}
        >
          {alert}
        </div>
        {handleOnClick && (
          <button
            className="rounded-lg px-10 py-1.5 bg-blue-800 absolute right-1 top-1"
            onClick={handleOnClick}
          >
            {"Download"}
          </button>
        )}
      </div>
    </div>
  );

  return [AlertComponent, TriggerAlert];
}
