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

  const getJobProgress = async (jobId) => {
    let response;

    await axios
      .post(urlBuilder("getSystemProgress"), { jobId })
      .then((value) => (response = value))
      .catch((err) => console.log({ err }));
    return response;
  };

  const submitJob = async (SubmitJobSwagger) => {
    let response;
    await axios
      .post(urlBuilder(`openWS/${123}`), {
        ...SubmitJobSwagger,
      })
      .then((value) => (response = value))
      .catch((err) => console.log("Error", err));

    return response;
  };

  const getAllFinishedJobs = async () => {
    let response;

    await axios
      .post(urlBuilder("getAllFinishedJobs"), {
        clientId: "123",
      })
      .then((value) => (response = value))
      .catch((err) => console.log("Error", err));

    return response;
  };

  return {
    submitJob,
    getAllBinaries,
    getJobProgress,
    uploadBinaries,
    getAllFinishedJobs,
    getAllFinishedJobs,
  };
};
