import React, { useContext, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import DropDownBox from "../components/DropDownBox";
import UploadFileButton from "../components/UploadFileButton.jsx";
import { AppContext } from "../context/AppContext.js";
import useAlert from "../helpers/useAlert.jsx";
import { getCompressedFile } from "../services/ServiceTypes/HandlerGroup.js";

const uuid = require("react-uuid");

export default function SubmitJob(props) {
  const { wsClient, binaries } = props;
  const navigate = useNavigate();

  const [AlertComponent, TriggerAlert] = useAlert();
  const [isSuccess, setIsSuccess] = useState(false);

  const { apiEndPoint, clientId } = useContext(AppContext);

  console.log({ binaries });

  const jobContentInput = useRef();

  const [isLoading, setIsLoading] = useState(false);

  const [distributeSelectedFile, setDistributeSelectedFile] = React.useState(
    JSON.parse(sessionStorage.getItem("distribute") ?? "{}")
  );
  const [processSelectedFile, setProcessSelectedFile] = React.useState(
    JSON.parse(sessionStorage.getItem("process") ?? "{}")
  );
  const [aggregateSelectedFile, setAggregateSelectedFile] = React.useState(
    JSON.parse(sessionStorage.getItem("aggregate") ?? "{}")
  );
  const [optionalFiles, setOptionalFiles] = React.useState({
    name: "",
    content: [],
  });

  console.log({ optionalFiles });

  const handleJobSubmit = async () => {
    setIsLoading(true);
    try {
      wsClient.sendMessage(
        `${JSON.stringify({
          jobId: uuid(),
          jobContent: jobContentInput.current.value,
          optionalFilesZip: optionalFiles,
          distributeBinaryId: distributeSelectedFile.id?.toString(),
          processBinaryId: processSelectedFile.id?.toString(),
          aggregateBinaryId: aggregateSelectedFile.id?.toString(),
        })}`
      );
      setIsSuccess(true);
      TriggerAlert("Submit Successed");
      navigate("/status");
    } catch (error) {
      console.log({ error });
      setIsSuccess(false);
      TriggerAlert("Submit Failed");
    }

    setIsLoading(false);
  };

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
            title={"distribute"}
            files={binaries.distribute}
            selectedFile={distributeSelectedFile}
            setSelectedFile={setDistributeSelectedFile}
          />
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
