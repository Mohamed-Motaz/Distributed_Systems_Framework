import { WebSocketServerService } from "../WebSocketServerService";

export const handleUploadFile = async (event, fileType, runCmd) => {
  const compressedFile = await getCompressedFile(event);

  const res = await WebSocketServerService().uploadBinaries(
    fileType,
    compressedFile.name,
    compressedFile.content,
    runCmd
  );

  return res;
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
  return res;
};

export const handleGetAllBinaries = async () => {
  return await WebSocketServerService().getAllBinaries();
};
