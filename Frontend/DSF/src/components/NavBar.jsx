/* eslint-disable react-hooks/exhaustive-deps */
import React, { useCallback, useEffect, useState } from 'react'
import { NavLink } from 'react-router-dom'
import DropDownMenu from './DropDownMenu';

import '../css/NavBarStyle.css'


export default function NavBar() 
{
    const [width, setWidth] = useState(null);
    const widthListener = useCallback(() => setWidth(window.innerWidth), [],)


    useEffect(() => {
        setWidth(window.innerWidth)
    }, [])

    useEffect(() => {
        window.addEventListener('resize', widthListener)
        return (() => {
            window.removeEventListener('resize', widthListener);
        })
    }, [])


    return <nav className='dark bg-black text-white shadow-md px-8 py-4  fixed top-0 left-0 right-0 z-50 flex items-center justify-between'>
        <NavLink className="nav-link"  to='/'>
            <h1 className='text-3xl font-bold logo'>DSF</h1>
        </NavLink>

        {
            width <= 700 ? <DropDownMenu />
            : <ul className='flex text-lg font-medium gap-x-4'>
                <li>
                    <NavLink className="nav-link" to='/'>Home</NavLink>
                </li>
                <li>
                    <NavLink className="nav-link" to='/events'>Events</NavLink>
                </li>
                <li>
                    <NavLink className="nav-link" to='/tops'>Tops</NavLink>
                </li>
                <li>
                    <NavLink className="nav-link" to='/about-us'>About Us</NavLink>
                </li>
            </ul>
        }

    </nav>

}
