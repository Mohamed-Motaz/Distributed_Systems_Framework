import UploadFileButton from "../components/UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import { WebSocketServerService } from "../services/WebSocketServerService";
import { Button } from "flowbite-react";
import { Buffer } from "buffer/";
import { Dropdown } from "@fluentui/react/lib/Dropdown";
import React from "react";
import TextField from "@material-ui/core/TextField";
import DropDownBox from "./DropDownBox";
import * as JSZip from "jszip";

//await blob.arrayBuffer().then((arrayBuffer) => Buffer.from(arrayBuffer, "binary"))
export const UploadFileButtons = (props) => {
  const { wsClient } = props;

  const [distribute, setDistribute] = React.useState([]);
  const [process, setProcess] = React.useState([]);
  const [aggregate, setAggregate] = React.useState([]);
  const [runCommand, setRunCommand] = React.useState("");
  console.log({ process });

  const [distributeSelectedFile, setDistributeSelectedFile] =
    React.useState("");
  const [processSelectedFile, setProcessSelectedFile] = React.useState("");
  const [aggregateSelectedFile, setAggregateSelectedFile] = React.useState("");

  const handleUploadFile = async (event, fileType) => {
    const fileUploaded = event.target.files[0];
    const zip = new JSZip();
    let base64Data;

    zip.file(fileUploaded.name, fileUploaded);
    console.log(fileUploaded);

    zip
      .generateAsync({ type: "blob", compression: "DEFLATE" })
      .then((content) => {
        console.log({ content });
        WebSocketServerService().uploadBinaries(
          fileType,
          fileUploaded.name,
          content,
          ""
        );
      });

    // zip.file(fileUploaded.name, fileUploaded);
    // let zipFileContent = await zip
    //   .generateAsync({ type: "blob" })
    //   .then((content) => content);

    // zipFileContent = await zipFileContent
    //   .arrayBuffer()
    //   .then((arrayBuffer) => Buffer.from(arrayBuffer, "binary"));

    // let arr = Array.from(Uint8Array.from(zipFileContent));

    // console.log({ arr });
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
        jobId: "123",
        clientId: "123",
        optionalFilesZip: {},
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

  return (
    <section className="m-8">
      <TextField
        labelName="Run command"
        color="secondary"
        value={runCommand}
        onChange={(cmd) => setRunCommand(cmd.target.value)}
      />
      <UploadFileButton
        onChange={(e) => handleUploadFile(e, BinariesType.process)}
        title={BinariesType.process}
      />
      <UploadFileButton
        onChange={(e) => handleUploadFile(e, BinariesType.Distribute)}
        title={BinariesType.Distribute}
      />
      <UploadFileButton
        onChange={(e) => handleUploadFile(e, BinariesType.aggregate)}
        title={BinariesType.aggregate}
      />
      <Button onClick={handleGetAllBinaries}>{"Get all Binaries"}</Button>
      <Button onClick={handleSubmitJob}>{"Submit job"}</Button>
      <Button onClick={getAllFinishedJob}>{"Get all finished jobs"}</Button>
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
    </section>
  );
};
