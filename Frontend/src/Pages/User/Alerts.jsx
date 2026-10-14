import { useState, useEffect } from "react";
import SideBar from "../Layout/SideBar";
import Header from "../Layout/Header";
import api from "../../API/api";
import { AlertTriangle, CheckCircle, Clock, RefreshCw } from "lucide-react";

const Alerts = () => {
    const [alertsList, setAlertsList] = useState([]);
    const [loading, setLoading] = useState(true);

    const fetchAlerts = async () => {
        try {
            setLoading(true);
            const response = await api.get("/alerts");
            if (Array.isArray(response.data)) {
                setAlertsList(response.data);
            }
        } catch (error) {
            console.error("Failed to fetch alerts:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchAlerts();
    }, []);

    const handleAcknowledge = (id) => {
        setAlertsList((prev) =>
            prev.map((a) => (a.id === id ? { ...a, status: "Resolved", severity: "RESOLVED" } : a))
        );
    };

    return (
        <div className="flex min-h-screen bg-slate-100 max-w-[2160px] mx-auto">
            <SideBar />
            <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
                <Header title="Incidents & Alerts" />

                <main className="flex-1 overflow-y-auto p-6 space-y-6">
                    <div>
                        <h1 className="text-2xl font-bold text-gray-800 tracking-tight">Active Incident Monitor</h1>
                        <p className="text-sm text-gray-500">Track outages, trigger alarms, and manage ongoing operational incidents.</p>
                    </div>

                    <div className="space-y-3">
                        {alertsList.length === 0 && !loading && (
                            <div className="bg-white p-12 rounded-2xl border border-gray-200/80 shadow-sm text-center">
                                <CheckCircle className="w-12 h-12 text-emerald-500 mx-auto mb-3" />
                                <h3 className="font-bold text-gray-800 text-lg">All Systems Operational</h3>
                                <p className="text-gray-400 text-sm mt-1">
                                    No active incidents or service outages reported on your endpoints.
                                </p>
                            </div>
                        )}
                        {alertsList.map((alert) => (
                            <div key={alert.id} className="bg-white p-5 rounded-2xl border border-gray-200/80 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:border-gray-300 transition">
                                <div className="flex items-center gap-4">
                                    <div className={`p-3 rounded-xl shrink-0 ${
                                        alert.severity === "CRITICAL" ? "bg-red-50 text-red-600" :
                                        alert.severity === "HIGH" ? "bg-amber-50 text-amber-600" :
                                        alert.severity === "MEDIUM" ? "bg-blue-50 text-blue-600" :
                                        "bg-emerald-50 text-emerald-600"
                                    }`}>
                                        {alert.status === "Resolved" ? <CheckCircle className="w-5 h-5" /> : <AlertTriangle className="w-5 h-5" />}
                                    </div>
                                    <div className="flex items-center gap-2.5 flex-wrap">
                                        <h4 className="font-bold text-gray-800 text-sm sm:text-base">{alert.title}</h4>
                                        <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                                            alert.severity === "CRITICAL" ? "bg-red-100 text-red-700" :
                                            alert.severity === "HIGH" ? "bg-amber-100 text-amber-700" :
                                            alert.severity === "MEDIUM" ? "bg-blue-100 text-blue-700" :
                                            "bg-emerald-100 text-emerald-700"
                                        }`}>
                                            {alert.severity}
                                        </span>
                                    </div>
                                </div>

                                <div className="flex items-center justify-between sm:justify-end gap-4 shrink-0 border-t sm:border-0 pt-3 sm:pt-0">
                                    <span className="text-xs text-gray-400 flex items-center gap-1 font-medium">
                                        <Clock className="w-3.5 h-3.5" /> {alert.time}
                                    </span>
                                    {alert.status !== "Resolved" ? (
                                        <button
                                            onClick={() => handleAcknowledge(alert.id)}
                                            className="bg-gray-100 hover:bg-[#002244] hover:text-white text-gray-700 font-semibold px-3 py-1.5 rounded-lg text-xs transition"
                                        >
                                            Need To Acknowledge
                                        </button>
                                    ) : (
                                        <span className="text-xs font-semibold text-emerald-600 bg-emerald-50 px-2.5 py-1 rounded-full">
                                            Closed
                                        </span>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>
                </main>
            </div>
        </div>
    );
};

export default Alerts;

