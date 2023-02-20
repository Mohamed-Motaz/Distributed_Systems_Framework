import { WebSocketServerService } from "../WebSocketServerService";

export const handleUploadFile = async (event, fileType, runCmd) => {
  const compressedFile = getCompressedFile(event);

  WebSocketServerService().uploadBinaries(
    fileType,
    compressedFile.name,
    compressedFile.content,
    runCmd
  );
};

export const getCompressedFile = async (event) => {
  const fileUploaded = event.target.files[0];

  const buffer = await fileUploaded.arrayBuffer();
  const view = new Uint8Array(buffer);
  console.log({ view });

  return { name: fileUploaded.name, content: Array.from(view) };
};

export const handleDeleteBinary = async (fileName, fileType) => {
  const res = await WebSocketServerService().deleteBinaryFile(
    fileName,
    fileType
  );
};

export const handleGetAllBinaries = async () => {
  return await WebSocketServerService().getAllBinaries();
};
