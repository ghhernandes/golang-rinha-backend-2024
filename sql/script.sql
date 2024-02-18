CREATE UNLOGGED TABLE clientes (
    id integer PRIMARY KEY,
    limite integer NOT NULL,
    saldo integer NOT NULL
);

INSERT INTO clientes (id, saldo, limite)
VALUES
  (1, 0, 1000 * 100),
  (2, 0, 800 * 100),
  (3, 0, 10000 * 100),
  (4, 0, 100000 * 100),
  (5, 0, 5000 * 100);

CREATE UNLOGGED TABLE transacoes (
    id SERIAL PRIMARY KEY,
    cliente_id int NOT NULL,
    valor integer NOT NULL,
    descricao text,
    data timestamp without time zone default (now() at time zone 'utc')
);

CREATE INDEX idx_transacoes_cliente_id ON transacoes (cliente_id ASC);
CREATE INDEX idx_transacoes_data ON transacoes (data DESC);

CREATE FUNCTION create_transacao(
    IN cliente_id integer, 
    IN valor integer, 
    IN descricao text,
    OUT novo_saldo integer,
    OUT limite_atual integer,
    OUT finalizado boolean
)
AS $$
BEGIN
    INSERT INTO transacoes(cliente_id, valor, descricao)
    VALUES (cliente_id, valor, descricao);

    UPDATE clientes SET
        saldo = saldo + valor
    WHERE id = cliente_id
    AND (valor > 0 OR saldo + valor >= (limite * -1))
    RETURNING limite, saldo
    INTO limite_atual, novo_saldo;

    finalizado = true;

    IF NOT FOUND THEN
        limite_atual = 0;
        novo_saldo = 0;
        finalizado = false;
    END IF;
END;
$$ LANGUAGE plpgsql;
