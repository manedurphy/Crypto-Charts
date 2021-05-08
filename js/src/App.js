import './App.css';
import { LineChart, Line, XAxis, YAxis } from 'recharts';
import { useEffect, useState } from 'react';

const url = process.env.NODE_ENV === 'development' ? 'http://localhost:8081/api/btc' : '/api/btc';

function App() {
    const [btcData, setBtcData] = useState([]);
    const [errMessage, setErrMessage] = useState('');

    useEffect(() => {
        fetch(url)
            .then((res) => res.json())
            .then((json) => (json.data ? setBtcData(json.data) : setErrMessage(json.message)));
    }, []);

    return btcData.length > 0 ? (
        <div>
            <h2>BTC Data</h2>
            <LineChart width={1850} height={400} data={btcData}>
                <XAxis dataKey="date" />
                <YAxis dataKey="value" />
                <Line type="monotone" dataKey="value" stroke="#8884d8" />
            </LineChart>
        </div>
    ) : (
        <div>{errMessage}</div>
    );
}

export default App;
