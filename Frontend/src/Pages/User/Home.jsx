import { useState, useEffect } from "react";
import SideBar from "../Layout/SideBar";
import Header from "../Layout/Header";
import api from "../../API/api";

const Home = () => {
    const [stats, setStats] = useState({
        total_nodes: 0,
        active_nodes: 0,
        healthy_nodes: 0,
        warning_nodes: 0,
        active_alerts: 0,
    });
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchDashboardData = async () => {
            try {
                // Load cached user if available
                const localUser = localStorage.getItem("user");
                if (localUser) {
                    try {
                        setUser(JSON.parse(localUser));
                    } catch (e) {}
                }

                // Fetch fresh profile and stats from backend
                const [statsRes, userRes] = await Promise.allSettled([
                    api.get("/dashboard/stats"),
                    api.get("/me"),
                ]);

                if (statsRes.status === "fulfilled" && statsRes.value.data) {
                    setStats(statsRes.value.data);
                }

                if (userRes.status === "fulfilled" && userRes.value.data?.user) {
                    setUser(userRes.value.data.user);
                    localStorage.setItem("user", JSON.stringify(userRes.value.data.user));
                }
            } catch (err) {
                console.error("Failed to load dashboard data:", err);
            } finally {
                setLoading(false);
            }
        };

        fetchDashboardData();
    }, []);

    const displayName = user?.name || user?.email?.split("@")[0] || "Operator";

    return (
        <div className="flex min-h-screen bg-slate-100 max-w-[2160px] mx-auto">
            <SideBar />

            <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <Header title="System Overview" user={user} />

                <main className="flex-1 overflow-y-auto p-6 space-y-6">
                    <div className="rounded-2xl p-6 text-[#002244] flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                        <div>
                            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight">
                                Welcome back, {displayName}! 👋
                            </h1>
                            <p className="text-sm text-gray-500 mt-1">
                                Real-time monitoring metrics from your PulseOps cluster.
                            </p>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm">
                            <span className="text-xs font-semibold text-gray-500">Total Monitored Nodes</span>
                            <div className="flex items-baseline justify-between mt-2">
                                <span className="text-2xl font-bold text-gray-900">
                                    {stats.total_nodes}
                                </span>
                            </div>
                        </div>

                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm">
                            <span className="text-xs font-semibold text-gray-500">Operational Nodes</span>
                            <div className="flex items-baseline justify-between mt-2">
                                <span className="text-2xl font-bold text-emerald-600">
                                    {stats.healthy_nodes}
                                </span>
                            </div>
                        </div>

                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm">
                            <span className="text-xs font-semibold text-gray-500">Degraded / Warnings</span>
                            <div className="flex items-baseline justify-between mt-2">
                                <span className="text-2xl font-bold text-amber-600">
                                    {stats.warning_nodes}
                                </span>
                            </div>
                        </div>

                        <div className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm">
                            <span className="text-xs font-semibold text-gray-500">Active Alerts</span>
                            <div className="flex items-baseline justify-between mt-2">
                                <span className="text-2xl font-bold text-red-600">
                                    {stats.active_alerts ?? 0}
                                </span>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        </div>
    );
};

export default Home;
