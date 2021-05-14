import './App.css';
import currencies from './data/currencies';
import BarComponent from './components/bar';
import LineComponent from './components/line';
import Title from './components/title';
import { handleDateConversion, getUrl } from './helpers/helpers';
import { Fragment, useEffect, useState } from 'react';

function App() {
    const [data, setData] = useState(null);
    const [current, setCurrent] = useState('bar');

    useEffect(() => {
        fetch(getUrl(current))
            .then((res) => res.json())
            .then((json) => (current !== 'bar' ? setData(handleDateConversion(json.Data.Data)) : setData(json.data)));
    }, [current]);

    return (
        data && (
            <Fragment>
                <Title currency={currencies[current]} />
                <div style={{ width: 1850, height: 700 }} className="container">
                    {current === 'bar' ? <BarComponent data={data} /> : <LineComponent data={data} />}
                </div>
                <div className="container">
                    <select onChange={(e) => setCurrent(e.target.value)} value={current}>
                        <option value="bar">All</option>
                        <option value="btc">Bitcoin</option>
                        <option value="eth">Ethereum</option>
                        <option value="doge">Doge</option>
                    </select>
                </div>
            </Fragment>
        )
    );
}

export default App;
