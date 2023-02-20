import { handleUploadFile } from "../services/ServiceTypes/HandlerGroup.js";
import React, { useRef } from "react";
import { FileTypeRadioButtons } from "./FileTypeRadioButtons";
import UploadFileButton from "./UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";

export const UploadFileCard = (props) => {
  const { handleGetAllBinaries } = props;
  const runCommandInput = useRef();
  const [fileType, setFileType] = React.useState(BinariesType.process);
  const [uploadedFileData, setUploadedFileData] = React.useState({
    event: {},
    runCmd: "",
  });

  return (
    <div className="flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
      <section className="mt-6 w-full ">
        <h3 className="md:text-2xl text-xl ">Run Command</h3>
        <textarea
          className="w-full rounded-lg border-2 border-blue-800 outline-none bg-black"
          ref={runCommandInput}
        />
      </section>

      <FileTypeRadioButtons fileType={fileType} setFileType={setFileType} />

      <div className="gap-5 flex flex-col">
        <UploadFileButton
          onChange={(e) =>
            setUploadedFileData({
              event: e,
              runCmd: runCommandInput.current.value,
            })
          }
          title={fileType}
        />
        <button
          className="rounded-lg px-10 py-1.5 bg-blue-800"
          onClick={() =>
            handleUploadFile(
              uploadedFileData.event,
              fileType,
              uploadedFileData.runCmd
            ).then((res) => {
              if (res.data.success) {
                handleGetAllBinaries();
              }
            })
          }
        >
          Upload
        </button>
      </div>
    </div>
  );
};
