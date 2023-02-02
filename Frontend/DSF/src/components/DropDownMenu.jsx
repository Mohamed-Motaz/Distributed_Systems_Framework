import { Dropdown } from 'flowbite-react'
import { NavLink } from 'react-router-dom';


export default function DropDownMenu() {


    return <Dropdown className='dark' label="Menu" style={{ background: '#1744e1' }}>

        <NavLink className="nav-link"  to='/'>
            <Dropdown.Item className='justify-center'>
                Home
            </Dropdown.Item>

        </NavLink>

        <NavLink className="nav-link"  to='/Loading'>
            <Dropdown.Item className='justify-center'>
                Events
            </Dropdown.Item>
        </NavLink >

        <NavLink className="nav-link"  to='/tops'>
            <Dropdown.Item className='justify-center'>
                Tops
            </Dropdown.Item>
        </NavLink>
        <NavLink className="nav-link"  to='/about-us'>
            <Dropdown.Item className='justify-center'>
                About Us
            </Dropdown.Item>
        </NavLink>
    </Dropdown>
}
