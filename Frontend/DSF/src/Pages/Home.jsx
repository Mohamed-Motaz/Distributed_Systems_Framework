import React from "react";
import useAlert from "../helpers/useAlert.jsx";

export default function Home() {
  const [AlertComponent, TriggerAlert] = useAlert();

  return (
    <>
      <main>
        <AlertComponent success={false} />

        <button
          className="m-16 p-4 rounded-lg bg-blue-800"
          onClick={() => TriggerAlert("My Alert Message")}
        >
          Trigger Alert
        </button>
      </main>
    </>
  );
}
