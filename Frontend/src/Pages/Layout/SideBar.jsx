import { Link, useLocation } from "react-router-dom";
import { LayoutDashboard, Server, Activity, AlertTriangle } from "lucide-react";

const SideBar = () => {
    const location = useLocation();

    const navItems = [
        {
            name: "Dashboard",
            path: "/home",
            icon: <LayoutDashboard className="w-5 h-5" />,
        },
        {
            name: "Services & Nodes",
            path: "/nodes",
            icon: <Server className="w-5 h-5" />,
        },
        {
            name: "Incidents & Alerts",
            path: "/alerts",
            icon: <AlertTriangle className="w-5 h-5" />,
        },
    ];

    return (
        <aside className="w-64 bg-[#001830] text-gray-300 min-h-screen flex flex-col border-r border-[#002244] shrink-0 sticky top-0 h-screen select-none">
            <div className="h-16 flex items-center px-5 border-b border-[#002d5a]">
                <Link to="/home" className="flex items-center gap-3">
                    <div className="flex flex-col">
                        <span className="font-extrabold text-white text-lg ">
                            Pulse<span className="text-cyan-400">Ops</span>
                        </span>
                    </div>
                </Link>
            </div>

            <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
                <div className="text-[12px] font-bold text-gray-400 px-3 mb-2">
                    Main Navigation
                </div>

                {navItems.map((item) => {
                    const isActive = location.pathname === item.path;
                    return (
                        <Link
                            key={item.path}
                            to={item.path}
                            className={`flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-colors group relative ${
                                isActive
                                    ? "bg-[#003366] text-white shadow-sm"
                                    : "text-gray-400 hover:bg-[#002244] hover:text-white"
                            }`}
                        >
                            <span className={`${isActive ? "text-cyan-400" : "text-gray-400 group-hover:text-cyan-300"} shrink-0`}>
                                {item.icon}
                            </span>
                            <span className="flex-1 truncate">{item.name}</span>
                        </Link>
                    );
                })}
            </nav>
        </aside>
    );
};

export default SideBar;

