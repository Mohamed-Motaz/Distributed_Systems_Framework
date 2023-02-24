import { set } from "lodash";
import React, { useContext, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { AppContext } from "../context/AppContext.js";
import DropDownBox from "../components/DropDownBox";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import { getCompressedFile } from "../services/ServiceTypes/HandlerGroup.js";
import UploadFileButton from "../components/UploadFileButton.jsx";
import { Button } from "flowbite-react";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import useAlert from "../helpers/useAlert.jsx";

const uuid = require("react-uuid");

export default function SubmitJob(props) {
  const { wsClient } = props;
  const navigate = useNavigate();

  const [AlertComponent, TriggerAlert] = useAlert();
  const [isSuccess, setIsSuccess] = useState(false);

  const { apiEndPoint, clientId, setClientId, setAllBinaries, binaries } =
    useContext(AppContext);

  const jobContentInput = useRef();

  const [isLoading, setIsLoading] = useState(false);

  const [distributeSelectedFile, setDistributeSelectedFile] =
    React.useState(null);
  const [processSelectedFile, setProcessSelectedFile] = React.useState(null);
  const [aggregateSelectedFile, setAggregateSelectedFile] =
    React.useState(null);
  const [optionalFiles, setOptionalFiles] = React.useState({
    name: "",
    content: [],
  });

  const handleJobSubmit = async () => {
    setIsLoading(true);
    try {
      wsClient.sendMessage(
        `${JSON.stringify({
          jobId: uuid(),
          clientId,
          jobContent: jobContentInput.current.value,
          optionalFilesZip: optionalFiles,
          distributeBinaryName: distributeSelectedFile,
          processBinaryName: processSelectedFile,
          aggregateBinaryName: aggregateSelectedFile,
        })}`
      );
      setIsSuccess(true);
      TriggerAlert("Submit Successed");
    } catch (error) {
      console.log({ error });
      setIsSuccess(false);
      TriggerAlert("Submit Failed");
    }

    setIsLoading(false);
    navigate("/status");
  };

  React.useEffect(() => {
    setAllBinaries();

    const intervalCalling = setInterval(async () => {
      //console.log("getJobsProgress() : Start...");
      await setAllBinaries();
      //console.log("getJobsProgress() : Done");
    }, 5000);

    return () => {
      clearInterval(intervalCalling);
    };
  }, []);

  const handleRandomizeClientId = () => setClientId(uuid());

  return (
    <main className="flex flex-col items-center pb-20 md:px-16">
      <AlertComponent success={isSuccess} />

      <h1 className="md:text-5xl text-3xl mb-8">Submit Job</h1>

      <div className="flex flex-col shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
        <section className="w-full flex items-center justify-start gap-2">
          <h3 className="md:text-2xl text-xl ">API Endpoint</h3>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{apiEndPoint || "http://localhost:5000/api/v1/test"}</p>
          </div>
        </section>

        <section className="w-full flex items-center justify-start gap-3">
          <h3 className="md:text-2xl text-xl ">Client ID</h3>
          <div className="w-fit rounded-lg border-2 border-blue-800 outline-none bg-black px-3 py-1">
            <p>{clientId}</p>
          </div>
          <button
            className="rounded-lg px-14 py-2 bg-blue-800  mt-8 justify-center text-xl"
            onClick={handleRandomizeClientId}
          >
            {"Randomize client id"}
          </button>
        </section>

        <section className="mt-6 w-full">
          <h3 className="md:text-2xl text-xl ">Job Content</h3>
          <textarea
            className="w-full rounded-lg border-2 border-blue-800 outline-none bg-black"
            ref={jobContentInput}
          />
        </section>

        <section className="flex gap-5 w-full justify-center mt-8">
          <DropDownBox
            title={"process"}
            files={binaries.process}
            selectedFile={processSelectedFile}
            setSelectedFile={setProcessSelectedFile}
          />
          <DropDownBox
            title={"aggregate"}
            files={binaries.aggregate}
            selectedFile={aggregateSelectedFile}
            setSelectedFile={setAggregateSelectedFile}
          />
          <DropDownBox
            title={"distribute"}
            files={binaries.distribute}
            selectedFile={distributeSelectedFile}
            setSelectedFile={setDistributeSelectedFile}
          />
        </section>
        <UploadFileButton
          onChange={async (e) => setOptionalFiles(await getCompressedFile(e))}
          title={"Optional"}
        />
        <button
          className="rounded-lg px-14 py-2 bg-blue-800 w-fit mt-8 self-center text-xl"
          onClick={handleJobSubmit}
        >
          {isLoading ? "Submitting..." : "Submit Job"}
        </button>
      </div>
    </main>
  );
}
