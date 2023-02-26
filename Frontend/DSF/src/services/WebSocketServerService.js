import axios from "axios";
import { urlBuilder } from "./ApiHelper";
import { BinariesType } from "./ServiceTypes/WebSocketServiceTypes";

export const WebSocketServerService = () => {
  const getAllBinaries = async () => {
    try {
      return await axios.post(urlBuilder("getAllBinaries"));
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };
  const uploadBinaries = async (fileType, name, content, runCmd) => {
    try {
      return await axios.post(urlBuilder("uploadBinary"), {
        fileType,
        name,
        content,
        runCmd,
      });
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };

  const deleteBinaryFile = async (fileName, fileType) => {
    try {
      return await axios.post(urlBuilder("deleteBinary"), {
        fileName,
        fileType,
      });
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };

  const getJobProgress = async (jobId) => {
    try {
      return await axios.post(urlBuilder("getSystemProgress"), { jobId });
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };

  const submitJob = async (SubmitJobSwagger, clientId) => {
    try {
      return await axios.post(urlBuilder(`openWS/${clientId}`), {
        ...SubmitJobSwagger,
      });
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };

  const getAllFinishedJobs = async (clientId) => {
    console.log({ clientId });
    try {
      return await axios.post(urlBuilder("getAllFinishedJobsIds"), {
        clientId,
      });
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };

  const pingEndPoint = async () => {
    try {
      await axios.get(urlBuilder("ping"));
      return true;
    } catch (_) {
      return false;
    }
  };

  const getJobById = async (clientId, jobId) => {
    try {
      return await axios.post(urlBuilder("getFinishedJobById"), {
        clientId,
        jobId,
      });
    } catch (err) {
      if (err.response?.data) {
        return err.response;
      } else {
        return null;
      }
    }
  };

  return {
    submitJob,
    getJobById,
    pingEndPoint,
    getAllBinaries,
    getJobProgress,
    uploadBinaries,
    deleteBinaryFile,
    getAllFinishedJobs,
  };
};
