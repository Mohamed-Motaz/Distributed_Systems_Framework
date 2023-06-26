import { Tooltip } from "flowbite-react";
import React from "react";
import { handleDeleteBinary } from "../services/ServiceTypes/HandlerGroup.js";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import { DeleteFileTypeRadioButtons } from "./DeleteFileTypeRadioButtons.jsx";
import DropDownBox from "./DropDownBox.jsx";

export const DeleteFileCard = (props) => {
  const { binaries, handleGetAllBinaries, TriggerAlert, setIsSuccess } = props;
  const [deleteFileType, setDeleteFileType] = React.useState(
    BinariesType.Distribute
  );

  const [selectedFile, setSelectedFile] = React.useState("");

  const getFilesByType =
    deleteFileType === BinariesType.process
      ? binaries.process
      : deleteFileType === BinariesType.Distribute
      ? binaries.distribute
      : binaries.aggregate;

  return (
    <div className="flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
      <h3 className="md:text-2xl text-xl ">Choose file to delete</h3>

      <DeleteFileTypeRadioButtons
        deleteFileType={deleteFileType}
        setDeleteFileType={setDeleteFileType}
      />
      <section className="flex gap-5 w-full justify-center  mt-8">
        <DropDownBox
          title={deleteFileType}
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
                deleteFileType,
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
