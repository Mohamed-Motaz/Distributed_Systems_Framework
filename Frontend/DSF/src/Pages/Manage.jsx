import React, { useContext, useRef, useState } from "react";
import { AppContext } from "../context/AppContext.js";
import { UploadFileCard } from "../components/UploadFileCard";
import { DeleteFileCard } from "../components/DeleteFileCard.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import Home from "./Home.jsx";
import { handleUploadFile } from "../services/ServiceTypes/HandlerGroup.js";
import UploadFileButton from "../components/UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import { FileTypeRadioButtons } from "./../components/FileTypeRadioButtons";
import useAlert from "../helpers/useAlert.jsx";
import { handleGetAllBinaries } from "../services/ServiceTypes/HandlerGroup.js";

export default function Manage() {
  const [isSubmittingApi, setIsSubmittingApi] = useState(false);
  const { changeApiEndPoint, apiEndPoint } = useContext(AppContext);

  const apiEndPointInput = useRef();

  const [AlertComponent, TriggerAlert] = useAlert();

  const [isSuccess, setIsSuccess] = React.useState(false);

  const handleOnClick = async () => {
    setIsSubmittingApi(true);
    setIsSuccess(false);
    const isAlive = await changeApiEndPoint(apiEndPointInput.current.value);
    apiEndPointInput.current.value = "";
    setIsSubmittingApi(false);
    if (!isAlive) {
      TriggerAlert("Endpoint is not found (Not Alive)");
      return;
    }
    setIsSuccess(true);
    TriggerAlert("Endpoint is set successfully");
  };

  const [binaries, setBinaries] = useState({
    process: [],
    aggregate: [],
    distribute: [],
  });

  const setAllBinaries = async (TriggerAlert, setIsSuccess) => {
    const files = await handleGetAllBinaries(TriggerAlert, setIsSuccess);

    const { AggregateBinaryNames, ProcessBinaryNames, DistributeBinaryNames } =
      files?.data?.response;
    setBinaries({
      process: ProcessBinaryNames,
      aggregate: AggregateBinaryNames,
      distribute: DistributeBinaryNames,
    });
  };

  React.useEffect(() => {
    setAllBinaries(TriggerAlert, setIsSuccess);

    const intervalCalling = setInterval(async () => {
      //console.log("getJobsProgress() : Start...");
      await setAllBinaries(TriggerAlert, setIsSuccess);
      //console.log("getJobsProgress() : Done");
    }, 5000);

    return () => {
      clearInterval(intervalCalling);
    };
  }, []);

  return (
    <main className="flex gap-5 flex-col items-center pb-20 md:px-16">
      <AlertComponent success={isSuccess} />

      <h1 className="md:text-5xl text-3xl mb-8">Manage</h1>

      <UploadFileCard setIsSuccess={setIsSuccess} TriggerAlert={TriggerAlert} />

      <DeleteFileCard
        binaries={binaries}
        setIsSuccess={setIsSuccess}
        TriggerAlert={TriggerAlert}
      />

      <div className="flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
        <section className="w-full flex items-center">
          <h3 className="md:text-2xl text-xl w-52">API Endpoint</h3>
          <div className="w-full relative">
            <input
              className="w-full rounded-lg border-2 border-blue-800 outline-none bg-black pr-36"
              type="text"
              placeholder="http://websocket-server"
              ref={apiEndPointInput}
            />
            <button
              className="rounded-lg px-10 py-1.5 bg-blue-800 absolute right-1 top-1"
              onClick={handleOnClick}
            >
              {isSubmittingApi ? "Connecting..." : "Connect"}
            </button>
          </div>
        </section>
      </div>
    </main>
  );
}
