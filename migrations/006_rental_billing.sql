BEGIN;

-- Accept payments outside the historical Pagila sample partitions.
-- Preserve the reporting views while allowing fees above the old $999.99 cap.
DROP MATERIALIZED VIEW rental_by_category;
DROP VIEW sales_by_film_category;
DROP VIEW sales_by_store;
ALTER TABLE payment ALTER COLUMN amount TYPE numeric(10,2);
CREATE MATERIALIZED VIEW public.rental_by_category AS
 SELECT c.name AS category,
    sum(p.amount) AS total_sales
   FROM (((((public.payment p
     JOIN public.rental r ON ((p.rental_id = r.rental_id)))
     JOIN public.inventory i ON ((r.inventory_id = i.inventory_id)))
     JOIN public.film f ON ((i.film_id = f.film_id)))
     JOIN public.film_category fc ON ((f.film_id = fc.film_id)))
     JOIN public.category c ON ((fc.category_id = c.category_id)))
  GROUP BY c.name
  ORDER BY (sum(p.amount)) DESC
  WITH NO DATA;

CREATE VIEW public.sales_by_film_category AS
 SELECT c.name AS category,
    sum(p.amount) AS total_sales
   FROM (((((public.payment p
     JOIN public.rental r ON ((p.rental_id = r.rental_id)))
     JOIN public.inventory i ON ((r.inventory_id = i.inventory_id)))
     JOIN public.film f ON ((i.film_id = f.film_id)))
     JOIN public.film_category fc ON ((f.film_id = fc.film_id)))
     JOIN public.category c ON ((fc.category_id = c.category_id)))
  GROUP BY c.name
  ORDER BY (sum(p.amount)) DESC;

CREATE VIEW public.sales_by_store AS
 SELECT ((c.city || ','::text) || cy.country) AS store,
    ((m.first_name || ' '::text) || m.last_name) AS manager,
    sum(p.amount) AS total_sales
   FROM (((((((public.payment p
     JOIN public.rental r ON ((p.rental_id = r.rental_id)))
     JOIN public.inventory i ON ((r.inventory_id = i.inventory_id)))
     JOIN public.store s ON ((i.store_id = s.store_id)))
     JOIN public.address a ON ((s.address_id = a.address_id)))
     JOIN public.city c ON ((a.city_id = c.city_id)))
     JOIN public.country cy ON ((c.country_id = cy.country_id)))
     JOIN public.staff m ON ((s.manager_staff_id = m.staff_id)))
  GROUP BY cy.country, c.city, s.store_id, m.first_name, m.last_name
  ORDER BY cy.country, c.city;
CREATE UNIQUE INDEX rental_category ON rental_by_category(category);
REFRESH MATERIALIZED VIEW rental_by_category;

CREATE TABLE payment_default PARTITION OF payment DEFAULT;
ALTER TABLE payment_default
    ADD CONSTRAINT payment_default_customer_fkey FOREIGN KEY (customer_id) REFERENCES customer(customer_id),
    ADD CONSTRAINT payment_default_staff_fkey FOREIGN KEY (staff_id) REFERENCES staff(staff_id),
    ADD CONSTRAINT payment_default_rental_fkey FOREIGN KEY (rental_id) REFERENCES rental(rental_id);
CREATE INDEX payment_default_customer_idx ON payment_default(customer_id);

-- One charge definition for both services: base rate plus $1 per started
-- 24-hour day overdue. Returned rentals stop accruing fees at return_date.
CREATE VIEW rental_charge AS
SELECT r.rental_id, r.customer_id,
       (f.rental_rate + GREATEST(
           CEIL(EXTRACT(EPOCH FROM (COALESCE(r.return_date, now()) - r.rental_date)) / 86400)
           - f.rental_duration, 0))::numeric AS amount
FROM rental r
JOIN inventory i ON i.inventory_id = r.inventory_id
JOIN film f ON f.film_id = i.film_id;

COMMIT;
