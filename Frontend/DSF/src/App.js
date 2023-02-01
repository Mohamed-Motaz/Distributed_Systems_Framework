import React, { useEffect } from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';

import RootLayout from './components/RootLayout';
import Home from './Pages/Home'
import NotFound from './Pages/NotFound'
import './App.css';

import Loading from './components/Loading.jsx';

const HOME_ROUTE = createBrowserRouter([
  {
    path: '/', element: <RootLayout/>, children: [
      { index: true, element: <Home /> },
      { path: '/Loading', element: <Loading /> },
      // {
      //   path: '/movies', element: <Outlet></Outlet>,
      //   children: [
      //     { index: true, element: <Movies /> },
      //     { path: 'details/:type/:id', element: <ItemDetails /> },
      //   ]
      // },
      { path: '*', element: <NotFound /> },
    ]
  }
]);


export default function App() 
{

  useEffect(() => {
    console.log('====================================');
    console.log("DID APP MOUNT");
    console.log('====================================');
  }, [])

    return <RouterProvider router={HOME_ROUTE} />

}
