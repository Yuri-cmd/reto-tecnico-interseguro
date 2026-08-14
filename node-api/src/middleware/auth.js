const jwt = require('jsonwebtoken');

/**
 * Middleware que valida el header "Authorization: Bearer <token>" contra
 * el mismo secreto compartido (JWT_SECRET) usado por la API de Go, tanto
 * para tokens emitidos al cliente como para la llamada interna Go -> Node.
 */
function verificarJWT(secret) {
  return (req, res, next) => {
    const authHeader = req.headers.authorization || '';
    if (!authHeader.startsWith('Bearer ')) {
      return res.status(401).json({ error: 'falta el header Authorization: Bearer <token>' });
    }

    const token = authHeader.slice('Bearer '.length);
    try {
      jwt.verify(token, secret);
      next();
    } catch (err) {
      return res.status(401).json({ error: 'token inválido o expirado' });
    }
  };
}

module.exports = { verificarJWT };
