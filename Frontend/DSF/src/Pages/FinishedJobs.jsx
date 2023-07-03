import { saveAs } from "file-saver";
import React, { useContext } from "react";
import { FaDownload } from "react-icons/fa";
import { AppContext } from "../context/AppContext.js";
import useAlert from "../helpers/useAlert.jsx";
import { WebSocketServerService } from "../services/WebSocketServerService.js";

export const handleDownloadJobById = async (
  TriggerAlert,
  setIsSuccess,
  jobId,
  clientId
) => {
  const job = await WebSocketServerService().getJobById(clientId, jobId);
  console.log({ job });
  if (!job?.data?.success) {
    setIsSuccess(false);
    TriggerAlert(
      job?.data?.response ??
        "Unable to establish the communication with the server"
    );
    return;
  }
  // downloadItem(
  //   job?.data?.response?.jobResult,
  //   `${job?.data?.response?.jobId}.txt`
  // );
  console.log("bhawel a download");
  const file = new Blob([job?.data?.response?.jobResult]);
  console.log({ file });
  saveAs(file, `${job?.data?.response?.jobId}.txt`);
  setIsSuccess(true);
  TriggerAlert("Job result downloaded successfully");
};

export default function FinishedJobs(props) {
  const { jobs } = props;

  const { clientId } = useContext(AppContext);
  const [AlertComponent, TriggerAlert] = useAlert();
  const [isSuccess, setIsSuccess] = React.useState(false);

  return (
    <main className="flex flex-col items-center pb-20 md:px-6">
      <AlertComponent success={isSuccess} />
      <h1 className="md:text-5xl text-3xl mb-16">Finished Jobs</h1>
      <section>
        {jobs?.length ? (
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
                      style={{ display: "initial" }}
                      onClick={() => {
                        handleDownloadJobById(
                          TriggerAlert,
                          setIsSuccess,
                          row,
                          clientId
                        );
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
