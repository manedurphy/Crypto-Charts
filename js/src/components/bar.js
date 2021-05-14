import { BarChart, Bar, Legend, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

function BarComponent({ data }) {
    return (
        <ResponsiveContainer width="95%" height="85%">
            <BarChart
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
                <XAxis dataKey="name" />
                <YAxis />
                <Legend />
                <Tooltip />
                <CartesianGrid strokeDasharray="3 3" />
                <Bar dataKey="USD" barSize={80} fill="#8884d8" label="USD" />
                <Bar dataKey="EUR" barSize={80} fill="#FF0000" label="EUR" />
            </BarChart>
        </ResponsiveContainer>
    );
}

export default BarComponent;
