import axios from "axios";
import { urlBuilder } from "./ApiHelper";
import { BinariesType } from "./ServiceTypes/WebSocketServiceTypes";

export const WebSocketServerService = () => {
  const getAllBinaries = async () => {
    let response;

    await axios
      .post(urlBuilder("getAllBinaries"))
      .then((value) => (response = value))
      .catch((err) => console.log("ERROR: ", err));

    return response;
  };
  /*
FileType string `json:"fileType"`
    Name     string `json:"name"`
    Content  []byte `json:"content"`
    RunCmd   string `json:"runCmd"`
     */
  const uploadBinaries = async (fileType, name, content, runCmd) => {
    let response;

    await axios
      .post(urlBuilder("uploadBinary"), {
        fileType,
        name,
        content,
        runCmd,
      })
      .then((value) => (response = value))
      .catch((err) => console.log({ err }));

    return response;
  };

  const getAllFinishedJobs = async (ClientId) => {
    let response;

    await axios
      .get(urlBuilder("getAllFinishedJobs"), {
        ClientId,
      })
      .then((value) => (response = value))
      .catch((err) => console.log("ERROR: ", err));

    return response;
  };

  return { getAllBinaries, getAllFinishedJobs, uploadBinaries };
};
