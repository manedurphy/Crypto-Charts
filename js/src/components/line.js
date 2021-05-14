import { LineChart, Line, Legend, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

function LineComponent({ data }) {
    return (
        <ResponsiveContainer width="95%" height="85%">
            <LineChart
                width={500}
                height={300}
                data={data}
                margin={{
                    top: 20,
                    right: 30,
                    left: 20,
                    bottom: 5,
                }}
            >
                <XAxis dataKey="time" />
                <YAxis dataKey="high" />
                <Legend />
                <Tooltip />
                <CartesianGrid strokeDasharray="3 3" />
                <Line dataKey="high" barSize={80} stroke="#FF0000" />
                <Line dataKey="low" barSize={80} stroke="#8884d8" />
            </LineChart>
        </ResponsiveContainer>
    );
}

export default LineComponent;
