import { useContext } from "react";
import { WebSocketServerService } from "../WebSocketServerService";

export const handleUploadFile = async (
  event,
  fileType,
  runCmd,
  TriggerAlert
) => {
  const compressedFile = await getCompressedFile(event);

  const res = await WebSocketServerService().uploadBinaries(
    fileType,
    compressedFile.name,
    compressedFile.content,
    runCmd
  );
  console.log({ res });

  if (res?.data?.success) {
    TriggerAlert("File has been uploaded successfully");
  } else {
    TriggerAlert(
      res?.data?.response ??
        "Unable to establish the communication with the server"
    );
  }

  return res;
};

export const getCompressedFile = async (event) => {
  const fileUploaded = event.target.files[0];

  const buffer = await fileUploaded.arrayBuffer();
  const view = new Uint8Array(buffer);
  console.log({ view });

  return { name: fileUploaded.name, content: Array.from(view) };
};

export const handleDeleteBinary = async (fileName, fileType, TriggerAlert) => {
  const res = await WebSocketServerService().deleteBinaryFile(
    fileName,
    fileType
  );
  if (res?.data?.success) {
    TriggerAlert("File has been deleted successfully");
  } else {
    TriggerAlert(
      res?.data?.response ??
        "Unable to establish the communication with the server"
    );
  }
  return res;
};

export const handleGetAllBinaries = async (TriggerAlert) => {
  const res = await WebSocketServerService().getAllBinaries();

  if (!res?.data?.success) {
    TriggerAlert(
      res?.data?.response ??
        "Unable to establish the communication with the server"
    );
  }

  return res;
};
