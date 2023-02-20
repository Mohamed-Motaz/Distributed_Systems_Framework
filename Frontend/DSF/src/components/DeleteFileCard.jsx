import { handleUploadFile } from "../services/ServiceTypes/HandlerGroup.js";
import React, { useRef } from "react";
import { FileTypeRadioButtons } from "./FileTypeRadioButtons";
import UploadFileButton from "./UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import DropDownBox from "./DropDownBox.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import { handleDeleteBinary } from "../services/ServiceTypes/HandlerGroup.js";

export const DeleteFileCard = (props) => {
  const { process, distribute, aggregate, handleGetAllBinaries } = props;
  const [fileType, setFileType] = React.useState(BinariesType.process);

  const [selectedFile, setSelectedFile] = React.useState("");
  console.log({ process, distribute, aggregate });
  const getFilesByType =
    fileType === BinariesType.process
      ? process
      : fileType === BinariesType.Distribute
      ? distribute
      : aggregate;
  console.log({ selectedFile });

  return (
    <div className="flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
      <h3 className="md:text-2xl text-xl ">Choose file to delete</h3>

      <FileTypeRadioButtons fileType={fileType} setFileType={setFileType} />
      <section className="flex gap-5 w-full justify-center  mt-8">
        <DropDownBox
          title={fileType}
          files={getFilesByType}
          selectedFile={selectedFile}
          setSelectedFile={setSelectedFile}
        />
        <button
          className="rounded-lg px-14 py-2 bg-blue-800 w-fit mt-8 self-center text-xl"
          onClick={() =>
            handleDeleteBinary(selectedFile, fileType).then((res) => {
              console.log({ res });
              if (res.data.success) {
                handleGetAllBinaries();
              }
            })
          }
        >
          Delete
        </button>
      </section>
    </div>
  );
};
