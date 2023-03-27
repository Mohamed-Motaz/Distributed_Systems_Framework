import React, { useEffect, useState, useContext } from "react";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import useWebSocket from "react-use-websocket";
import RootLayout from "./components/RootLayout";
import Home from "./Pages/Home";
import NotFound from "./Pages/NotFound";
import Landing from "./Pages/Landing.jsx";
import HowTo from "./Pages/HowTo.jsx";
import { AppContext } from "../src/context/AppContext";
import Manage from "./Pages/Manage.jsx";
import SubmitJob from "./Pages/SubmitJob.jsx";
import useAlert from "./helpers/useAlert";
import "./App.css";
import Status from "./Pages/Status.jsx";
import AboutUs from "./Pages/AboutUs.jsx";
import FinishedJobs from "./Pages/FinishedJobs.jsx";

export default function App() {
  const [isFirst, setIsFirst] = useState(true);

  const { clientId, apiEndPoint } = useContext(AppContext);
  const WS_URL = `ws://${apiEndPoint}/openWS/${clientId}`;

  const [AlertComponent, TriggerAlert] = useAlert();

  const [isSuccess, setIsSuccess] = React.useState(false);

  const [binaries, setBinaries] = useState({
    process: [],
    aggregate: [],
    distribute: [],
  });

  const [finishedJobIds, setFinishedJobIds] = useState(null);
  const [systemProgress, setSystemProgress] = useState(null);

  const setAllBinaries = async (files) => {
    const { AggregateBinaryNames, ProcessBinaryNames, DistributeBinaryNames } =
      files;
    setBinaries({
      process: ProcessBinaryNames,
      aggregate: AggregateBinaryNames,
      distribute: DistributeBinaryNames,
    });
  };

  //getAllBinaries and getSystemProgress and finishedJobsIds
  console.log("API Endpoint =======>>>> ", apiEndPoint);
  const wsClient = useWebSocket(WS_URL, {
    onOpen: () => {
      console.log("WebSocket connection established.");
    },
    shouldReconnect: (closeEvent) => true,
    // onClose: () => {
    //   console.log(
    //     "WebSocket connection closed, it will be re-established in a second"
    //   );
    //   setTimeout(wsClient, 1000);
    // },
    onMessage: (e) => {
      if (e.data.msgType === "systemBinaries") {
        if (e.data.response.success) {
          setAllBinaries(e.data.response.response);
        } else {
          TriggerAlert(
            e?.data?.response.response ?? "Unable to get system binaries"
          );
        }
      } else if (e.data.msgType === "finishedJobsIds") {
        if (e.data.response.success) {
          setFinishedJobIds(e?.data?.response.response || []);
        } else {
          TriggerAlert(
            e?.data?.response.response ?? "Unable to get finished job IDs"
          );
        }
      } else if (e.data.msgType === "systemProgress") {
        if (e.data.response.success) {
          setSystemProgress(e?.data?.response.response || []);
        } else {
          TriggerAlert(
            e?.data?.response.response ?? "Unable to get systemProgress"
          );
        }
      } else if (e.data.msgType === "finishedJob") {
        if (e.data.response.success) {
          TriggerAlert(
            `The job with id: ${e.data.response.response.JobId} is finished`,
            () => {}
          ); // implement download logic
        } else {
          TriggerAlert(
            e?.data?.response.response ?? "Unable to get finished job"
          );
        }
      }

      console.log({ e });
    },
  });

  const HOME_ROUTE = createBrowserRouter([
    {
      path: "/",
      element: <RootLayout />,
      children: [
        { index: true, element: <Landing /> },
        { path: "/how-to", element: <HowTo /> },
        { path: "/manage", element: <Manage binaries={binaries} /> },
        {
          path: "/submit-job",
          element: <SubmitJob binaries={binaries} wsClient={wsClient} />,
        },
        { path: "/status", element: <Status jobs={systemProgress} /> },
        {
          path: "/finished-jobs",
          element: <FinishedJobs jobs={finishedJobIds} />,
        },
        { path: "/about-us", element: <AboutUs /> },
        // {
        //   path: '/movies', element: <Outlet></Outlet>,
        //   children: [
        //     { index: true, element: <Movies /> },
        //     { path: 'details/:type/:id', element: <ItemDetails /> },
        //   ]
        // },
        { path: "*", element: <NotFound /> },
      ],
    },
  ]);

  useEffect(() => {
    const result = localStorage.getItem("isFirst");
    if (result) {
      setIsFirst(false);
      localStorage.setItem("isFirst", "False");
    } else {
      localStorage.setItem("isFirst", "True");
      setIsFirst(true);
    }

    console.log("====================================");
    console.log("DID APP MOUNT");
    console.log("====================================");
  }, []);

  return (
    <main>
      <AlertComponent success={isSuccess} />

      <RouterProvider
        router={
          localStorage.getItem("apiEndPoint")
            ? HOME_ROUTE
            : createBrowserRouter([
                {
                  path: "/",
                  element: (
                    <div className="dark pt-28 px-8">
                      <Landing />
                    </div>
                  ),
                },
              ])
        }
      />
    </main>
  );
}
