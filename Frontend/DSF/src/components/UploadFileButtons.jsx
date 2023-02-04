import UploadFileButton from "../components/UploadFileButton.jsx";
import { BinariesType } from "../services/ServiceTypes/WebSocketServiceTypes.js";
import { WebSocketServerService } from "../services/WebSocketServerService";
import { Button } from "flowbite-react";
import { Buffer } from "buffer/";

//await blob.arrayBuffer().then((arrayBuffer) => Buffer.from(arrayBuffer, "binary"))
export const UploadFileButtons = () => {
  const handleUploadFile = async (event, fileType) => {
    const fileUploaded = event.target.files[0];
    const zip = require("jszip")();

    let zipFileContent = zip.file(fileUploaded.name, fileUploaded);
    zipFileContent = await zip
      .generateAsync({ type: "blob" })
      .then((content) => content);

    zipFileContent = await zipFileContent
      .arrayBuffer()
      .then((arrayBuffer) => Buffer.from(arrayBuffer, "binary"));

    let arr = Array.from(Uint8Array.from(zipFileContent));

    console.log({ arr });

    WebSocketServerService().uploadBinaries(
      fileType,
      fileUploaded.name,
      arr,
      ""
    );
  };

  const handleGetAllBinaries = async () => {
    const res = await WebSocketServerService().getAllBinaries();
    console.log({ Binaries: res });
  };

  return (
    <section className="m-8">
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
    </section>
  );
};
