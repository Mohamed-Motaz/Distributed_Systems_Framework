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
  console.log({ process });

  const [distributeSelectedFile, setDistributeSelectedFile] =
    React.useState("");
  const [processSelectedFile, setProcessSelectedFile] = React.useState("");
  const [aggregateSelectedFile, setAggregateSelectedFile] = React.useState("");

  const handleGetAllBinaries = async () => {
    const files = await WebSocketServerService().getAllBinaries();
    setAggregate(files.data.response.AggregateBinaryNames);
    setProcess(files.data.response.ProcessBinaryNames);
    setDistribute(files.data.response.DistributeBinaryNames);

    console.log({ Binaries: files });
  };

  const handleDeleteBinary = async (fileName, fileType) => {
    const res = await WebSocketServerService().deleteBinaryFile(
      fileName,
      fileType
    );
  };
  return (
    <section className="m-8">
      <Button onClick={handleGetAllBinaries}>{"Get all Binaries"}</Button>

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
    </section>
  );
};
