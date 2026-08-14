const express = require('express');
const cors = require('cors');

const { verificarJWT } = require('./middleware/auth');
const estadisticasRouter = require('./routes/estadisticas');

const PORT = process.env.PORT || 4000;
const JWT_SECRET = process.env.JWT_SECRET || 'dev-secret-cambiar-en-produccion';

const app = express();
app.use(cors());
app.use(express.json());

app.get('/health', (req, res) => res.json({ status: 'ok' }));

app.use('/api/estadisticas', verificarJWT(JWT_SECRET), estadisticasRouter);

app.listen(PORT, () => {
  console.log(`Node/Express API (estadísticas) escuchando en :${PORT}`);
});

module.exports = app;
