import { handleUploadFile } from "../services/ServiceTypes/HandlerGroup.js";
import React, { useRef, useContext } from "react";
import { FileTypeRadioButtons } from "./FileTypeRadioButtons";
import UploadFileButton from "./UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import DropDownBox from "./DropDownBox.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";
import { handleDeleteBinary } from "../services/ServiceTypes/HandlerGroup.js";
import { Tooltip } from "flowbite-react";

export const DeleteFileCard = (props) => {
  const { binaries, handleGetAllBinaries, TriggerAlert, setIsSuccess } = props;
  const [fileType, setFileType] = React.useState(BinariesType.process);

  const [selectedFile, setSelectedFile] = React.useState("");

  const getFilesByType =
    fileType === BinariesType.process
      ? binaries.process
      : fileType === BinariesType.Distribute
      ? binaries.distribute
      : binaries.aggregate;

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
        <Tooltip
          content={
            <h2>
              {!selectedFile
                ? "Please choose file to Delete"
                : "Click to delete"}
            </h2>
          }
        >
          <button
            className={`rounded-lg px-14 py-2 ${
              !selectedFile ? "bg-blue-800 opacity-60" : "bg-blue-800"
            } w-fit mt-8 self-center text-xl`}
            disabled={!selectedFile}
            onClick={() =>
              handleDeleteBinary(
                selectedFile,
                fileType,
                TriggerAlert,
                setIsSuccess
              ).then((res) => {
                console.log({ res });
                if (res.data.success) {
                  handleGetAllBinaries();
                }
              })
            }
          >
            Delete
          </button>
        </Tooltip>
      </section>
    </div>
  );
};
