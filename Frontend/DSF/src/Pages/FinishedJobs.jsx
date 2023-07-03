import React, { useContext } from "react";
import { FaDownload } from "react-icons/fa";
import { AppContext } from "../context/AppContext.js";
import { downloadItem } from "../helpers/index.js";
import { WebSocketServerService } from "../services/WebSocketServerService.js";

export default function FinishedJobs(props) {
  const { jobs } = props;

  const { TriggerAlert, clientId } = useContext(AppContext);

  const handleDownloadJobById = async (jobId) => {
    const job = await WebSocketServerService().getJobById(clientId, jobId);
    if (!job?.data?.success) {
      TriggerAlert(
        job?.data?.response ??
          "Unable to establish the communication with the server"
      );
      return;
    }
    downloadItem(
      job?.data?.response?.jobResult,
      `${job?.data?.response?.jobId}.txt`
    );
    TriggerAlert("Job result downloaded successfully");
  };

  return (
    <main className="flex flex-col items-center pb-20 md:px-6">
      <h1 className="md:text-5xl text-3xl mb-16">Finished Jobs</h1>
      <section>
        {jobs.length ? (
          <table className="w-full table-fixed">
            <thead>
              <tr>
                <th className="pb-5 text-xl">Job ID</th>
                <th className="pb-5 text-xl">Download</th>
              </tr>
            </thead>
            <tbody>
              {jobs?.map((row, i) => (
                <tr
                  className={`border-t-2 border-b-2 ${
                    i % 2 === 0 && "bg-white bg-opacity-10"
                  }`}
                >
                  <td className="text-center">{row}</td>

                  <td className="text-center">
                    <FaDownload
                      onClick={() => {
                        handleDownloadJobById(row);
                      }}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <h2 className="text-2xl self-start">No Finished Jobs Found</h2>
        )}
      </section>
    </main>
  );
}
