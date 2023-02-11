import UploadFileButton from "../components/UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import { WebSocketServerService } from "../services/WebSocketServerService";
import { Button } from "flowbite-react";
import React from "react";
import TextField from "@material-ui/core/TextField";
import DropDownBox from "./DropDownBox";
import uuid from "react-uuid";

//await blob.arrayBuffer().then((arrayBuffer) => Buffer.from(arrayBuffer, "binary"))
export const UploadFileButtons = (props) => {
  const { wsClient } = props;

  const [distribute, setDistribute] = React.useState([]);
  const [process, setProcess] = React.useState([]);
  const [aggregate, setAggregate] = React.useState([]);
  const [runCommand, setRunCommand] = React.useState("");
  const [jobContent, setJobContent] = React.useState("");
  console.log({ process });

  const [distributeSelectedFile, setDistributeSelectedFile] =
    React.useState("");
  const [processSelectedFile, setProcessSelectedFile] = React.useState("");
  const [aggregateSelectedFile, setAggregateSelectedFile] = React.useState("");
  const [optionalFiles, setOptionalFiles] = React.useState({
    name: "",
    content: [],
  });

  const handleUploadFile = async (event, fileType, runCmd) => {
    const fileUploaded = event.target.files[0];

    const buffer = await fileUploaded.arrayBuffer();
    const view = new Uint8Array(buffer);

    if (fileType === BinariesType.optionalFiles) {
      setOptionalFiles({ name: fileUploaded.name, content: Array.from(view) });
    }
    console.log({ view });

    WebSocketServerService().uploadBinaries(
      fileType,
      fileUploaded.name,
      Array.from(view),
      runCmd
    );
  };

  const handleGetAllBinaries = async () => {
    const files = await WebSocketServerService().getAllBinaries();
    setAggregate(files.data.response.AggregateBinaryNames);
    setProcess(files.data.response.ProcessBinaryNames);
    setDistribute(files.data.response.DistributeBinaryNames);

    console.log({ Binaries: files });
  };

  const handleSubmitJob = () => {
    console.log("get called");
    wsClient.sendMessage(
      `${JSON.stringify({
        jobId: uuid(),
        clientId: "123",
        jobContent,
        optionalFilesZip: optionalFiles,
        distributeBinaryName: distributeSelectedFile,
        processBinaryName: processSelectedFile,
        aggregateBinaryName: aggregateSelectedFile,
      })}`
    );
  };

  const getAllFinishedJob = async () => {
    const finishedJobs = await WebSocketServerService().getAllFinishedJobs();
    console.log({ finishedJobs });
  };

  const getJobProgress = async () => {
    const jobProgress = await WebSocketServerService().getJobProgress("123");

    console.log({ jobProgress });
  };

  const handleDeleteBinary = async (fileName, fileType) => {
    const res = await WebSocketServerService().deleteBinaryFile(
      fileName,
      fileType
    );
  };
  return (
    <section className="m-8">
      <TextField
        labelName="Run command"
        color="secondary"
        value={runCommand}
        onChange={(cmd) => setRunCommand(cmd.target.value)}
      />

      <UploadFileButton
        onChange={(e) => handleUploadFile(e, BinariesType.process, runCommand)}
        title={BinariesType.process}
      />
      <UploadFileButton
        onChange={(e) =>
          handleUploadFile(e, BinariesType.Distribute, runCommand)
        }
        title={BinariesType.Distribute}
      />
      <UploadFileButton
        onChange={(e) =>
          handleUploadFile(e, BinariesType.aggregate, runCommand)
        }
        title={BinariesType.aggregate}
      />
      <UploadFileButton
        onChange={(e) =>
          handleUploadFile(e, BinariesType.optionalFiles, runCommand)
        }
        title={BinariesType.optionalFiles}
      />

      <Button onClick={handleGetAllBinaries}>{"Get all Binaries"}</Button>
      <Button onClick={handleSubmitJob}>{"Submit job"}</Button>
      <Button onClick={getAllFinishedJob}>{"Get all finished jobs"}</Button>
      <Button onClick={getJobProgress}>{"Get Job Progress"}</Button>
      <Button
        onClick={() =>
          handleDeleteBinary(processSelectedFile, BinariesType.process)
        }
      >
        {"Delete process File"}
      </Button>
      <Button
        onClick={() =>
          handleDeleteBinary(aggregateSelectedFile, BinariesType.aggregate)
        }
      >
        {"Delete aggregate File"}
      </Button>
      <Button
        onClick={() =>
          handleDeleteBinary(distributeSelectedFile, BinariesType.Distribute)
        }
      >
        {"Delete distribute File"}
      </Button>
      <DropDownBox
        title={"process"}
        files={process}
        selectedFile={processSelectedFile}
        setSelectedFile={setProcessSelectedFile}
      />
      <DropDownBox
        title={"aggregate"}
        files={aggregate}
        selectedFile={aggregateSelectedFile}
        setSelectedFile={setAggregateSelectedFile}
      />
      <DropDownBox
        title={"distribute"}
        files={distribute}
        selectedFile={distributeSelectedFile}
        setSelectedFile={setDistributeSelectedFile}
      />

      <TextField
        labelName="Run command"
        color="secondary"
        value={jobContent}
        onChange={(cmd) => setJobContent(cmd.target.value)}
        multiline
      />
    </section>
  );
};
