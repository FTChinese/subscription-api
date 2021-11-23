package pw

const StmtCreateProduct = `
INSERT INTO subs_product.product
SET id = :product_id,
    tier = :tier,
    heading = :heading,
    description = :description,
    small_print = :small_print,
    created_by = :created_by,
    created_utc = :created_utc,
    updated_utc = :updated_utc
`
