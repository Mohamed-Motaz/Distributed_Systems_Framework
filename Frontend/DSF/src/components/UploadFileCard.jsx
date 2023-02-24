import { handleUploadFile } from "../services/ServiceTypes/HandlerGroup.js";
import React, { useRef, useContext } from "react";
import { FileTypeRadioButtons } from "./FileTypeRadioButtons";
import UploadFileButton from "./UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import { Tooltip } from "flowbite-react";
import { AppContext } from "../context/AppContext.js";

export const UploadFileCard = (props) => {
  const {} = props;
  const runCommandInput = useRef();

  const { TriggerAlert, setIsSuccess } = useContext(AppContext);
  const [runCmd, setRunCmd] = React.useState("");
  const [fileType, setFileType] = React.useState(BinariesType.process);
  const [uploadedFileData, setUploadedFileData] = React.useState({
    event: {},
  });

  const uploadButtonDisabled =
    runCmd.length === 0 || Object.keys(uploadedFileData.event).length === 0;

  React.useEffect(
    () =>
      setUploadedFileData({
        event: {},
      }),
    [fileType]
  );

  console.log({ uploadedFileData });
  const handleUpload = async () => {
    await handleUploadFile(
      uploadedFileData.event,
      fileType,
      runCmd,
      TriggerAlert,
      setIsSuccess
    );
  };

  return (
    <div className="flex flex-col justify-center items-center shadow-card hover:shadow-cardhover rounded-lg px-8 py-12 gap-2  w-full">
      <section className="mt-6 w-full ">
        <h3 className="md:text-2xl text-xl ">Run Command</h3>
        <textarea
          className="w-full rounded-lg border-2 border-blue-800 outline-none bg-black"
          onChange={(e) => setRunCmd(e.target.value)}
          ref={runCommandInput}
        />
      </section>

      <FileTypeRadioButtons fileType={fileType} setFileType={setFileType} />

      <div className="gap-5 flex flex-col">
        <UploadFileButton
          onChange={(e) =>
            setUploadedFileData({
              event: e,
            })
          }
          title={fileType}
        />
        <Tooltip
          content={
            <h2>
              {uploadButtonDisabled
                ? "Please check that you enter a run command and upload a file first"
                : "Click to upload"}
            </h2>
          }
        >
          <button
            className={`rounded-lg px-10 py-1.5 ${
              uploadButtonDisabled ? "bg-blue-800 opacity-60" : "bg-blue-800"
            }`}
            disabled={uploadButtonDisabled}
            onClick={handleUpload}
          >
            Upload
          </button>
        </Tooltip>
      </div>
    </div>
  );
};
