package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	pb "test/proto" // замените на ваш импорт
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	totalOrders    = 100000
	testDuration   = 60 * time.Minute
	baseRequestID  = "987666321"
)


var (
	products = []struct {
		ID    string
		Name  string
		Price int32
	}{
		{"prod1", "Хлеб", 50},
		{"prod2", "Молоко", 50},
		{"prod3", "Яйца", 80},
		{"prod4", "Сахар", 60},
		{"prod5", "Масло", 120},
		{"prod6", "Сыр", 200},
		{"prod7", "Колбаса", 180},
		{"prod8", "Чай", 90},
		{"prod9", "Кофе", 250},
		{"prod10", "Печенье", 70},
		{"prod11", "Макароны", 55},
		{"prod12", "Рис", 75},
		{"prod13", "Гречка", 65},
		{"prod14", "Мука", 45},
		{"prod15", "Соль", 20},
		{"prod16", "Перец", 30},
		{"prod17", "Лавровый лист", 25},
		{"prod18", "Томатная паста", 40},
		{"prod19", "Кетчуп", 60},
		{"prod20", "Майонез", 70},
		{"prod21", "Сметана", 65},
		{"prod22", "Творог", 90},
		{"prod23", "Йогурт", 45},
		{"prod24", "Кефир", 50},
		{"prod25", "Ряженка", 55},
		{"prod26", "Сгущенка", 85},
		{"prod27", "Мармелад", 95},
		{"prod28", "Шоколад", 110},
		{"prod29", "Конфеты", 130},
		{"prod30", "Вафли", 75},
		{"prod31", "Пряники", 65},
		{"prod32", "Зефир", 85},
		{"prod33", "Халва", 90},
		{"prod34", "Орехи", 150},
		{"prod35", "Семечки", 40},
		{"prod36", "Сухофрукты", 120},
		{"prod37", "Чипсы", 85},
		{"prod38", "Сушки", 35},
		{"prod39", "Сухари", 30},
		{"prod40", "Попкорн", 55},
		{"prod41", "Лимонад", 60},
		{"prod42", "Сок", 80},
		{"prod43", "Минералка", 35},
		{"prod44", "Пиво", 90},
		{"prod45", "Вино", 250},
		{"prod46", "Водка", 300},
		{"prod47", "Коньяк", 500},
		{"prod48", "Шампанское", 350},
		{"prod49", "Виски", 600},
		{"prod50", "Ром", 450},
		{"prod51", "Текила", 550},
		{"prod52", "Джин", 400},
		{"prod53", "Ликер", 350},
		{"prod54", "Вермут", 300},
		{"prod55", "Бренди", 400},
		{"prod56", "Самогон", 200},
		{"prod57", "Абсент", 700},
		{"prod58", "Пиво безалкогольное", 60},
		{"prod59", "Квас", 45},
		{"prod60", "Морс", 55},
		{"prod61", "Компот", 40},
		{"prod62", "Кисель", 35},
		{"prod63", "Смузи", 120},
		{"prod64", "Коктейль", 150},
		{"prod65", "Энергетик", 90},
		{"prod66", "Кола", 65},
		{"prod67", "Пепси", 65},
		{"prod68", "Фанта", 65},
		{"prod69", "Спрайт", 65},
		{"prod70", "Миринда", 65},
		{"prod71", "Байкал", 60},
		{"prod72", "Тархун", 60},
		{"prod73", "Дюшес", 60},
		{"prod74", "Буратино", 60},
		{"prod75", "Саяны", 60},
		{"prod76", "Лимон", 40},
		{"prod77", "Апельсин", 60},
		{"prod78", "Яблоко", 45},
		{"prod79", "Груша", 55},
		{"prod80", "Банан", 50},
		{"prod81", "Киви", 70},
		{"prod82", "Ананас", 120},
		{"prod83", "Манго", 150},
		{"prod84", "Персик", 80},
		{"prod85", "Нектарин", 85},
		{"prod86", "Слива", 60},
		{"prod87", "Абрикос", 70},
		{"prod88", "Вишня", 90},
		{"prod89", "Черешня", 110},
		{"prod90", "Клубника", 130},
		{"prod91", "Малина", 140},
		{"prod92", "Смородина", 100},
		{"prod93", "Крыжовник", 90},
		{"prod94", "Облепиха", 120},
		{"prod95", "Клюква", 110},
		{"prod96", "Брусника", 100},
		{"prod97", "Черника", 130},
		{"prod98", "Голубика", 150},
		{"prod99", "Ежевика", 140},
		{"prod100", "Арбуз", 200},
	}
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient("localhost:8080", opts...)
	if err != nil {
		log.Fatalf("Failed to conn: %v", err)
	}

	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	// Канал для контроля скорости
	rateLimiter := time.Tick(time.Second)
	
	// Счетчики
	var (
		successCount int64
		wg          sync.WaitGroup
	)

	startTime := time.Now()
	endTime := startTime.Add(testDuration)

	fmt.Printf("Starting test: %d orders in %v\n", totalOrders, testDuration)

	// Горутина для вывода прогресса
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime)
				rate := float64(successCount) / elapsed.Seconds()
				fmt.Printf("Progress: %d/%d (%.2f req/sec), remaining time: %v\n", 
					successCount, totalOrders, rate, endTime.Sub(time.Now()))
			}
		}
	}()

	// Основной цикл отправки запросов
	for i := 0; i < totalOrders; {
		// Проверяем, не истекло ли время теста
		if time.Now().After(endTime) {
			fmt.Println("Test duration exceeded, stopping...")
			break
		}

		// Ожидаем следующий тик для контроля скорости
		<-rateLimiter

		// Генерируем случайное количество запросов в эту секунду (1-10)
		requestsThisSecond := 20
		
		for j := 0; j < requestsThisSecond && i < totalOrders; j++ {
			wg.Add(1)
			deliveryTime := generateRandomDeliveryTime()
			go func(orderNum int, deliveryTime string) {

				defer wg.Done()
				// Генерируем уникальные параметры заказа
				requestID := fmt.Sprintf("%s-%05d6152s22sda22", baseRequestID, orderNum)
				items := generateRandomItems()
				totalPrice := int32(calculateTotalPrice(items))
				// Создаем запрос
				createReq := &pb.CreateOrderRequest{
					RequestId:      requestID,
					UserId:        "987666asds321",
					DarkstoreId:   "store789",
					CustomerName:    "Иван",
					CustomerSurname: "Иванов",
					CustomerPhone:   "+79161234567",
					DeliveryWindow:    deliveryTime,
					Address:          "ул. Пушкина, д. Колотушкина, кв. 123",
					UnderDoor:        rand.Intn(2) == 1,
					CallBefore:       rand.Intn(2) == 1,
					CommentToCourier: generateRandomComment(),
					TotalPrice:       totalPrice,
					Items:           items,
				}

				// Отправляем запрос
				ctx := context.Background()
				
				_, err := client.CreateOrder(ctx, createReq)
				if err != nil {
					fmt.Printf("Error creating order %s: %v\n", requestID, err)
					return
				}
				
				// Увеличиваем счетчик успешных запросов
				successCount++
			}(i, deliveryTime)
			
			i++ // Увеличиваем счетчик основного цикла
		}
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Printf("Test completed: %d orders in %v (%.2f req/sec)\n", 
		successCount, elapsed, float64(successCount)/elapsed.Seconds())
}

// Генерирует случайное время доставки в формате "HH:MM-HH:MM"
func generateRandomDeliveryTime() string {
	now := time.Now() 
	
	// Генерируем случайное время позже текущего момента:
	// - от +1 часа до +24 часов от текущего времени
	hoursToAdd := rand.Intn(23) + 1  // 1-23 часа
	minutesToAdd := rand.Intn(60)    // 0-59 минут
	
	deliveryTime := now.Add(time.Duration(hoursToAdd)*time.Hour + 
	                       time.Duration(minutesToAdd)*time.Minute)
	
	return fmt.Sprintf("%d:%02d:%02d:%02d:%02d",
		deliveryTime.Year()-1,
		deliveryTime.Month(),
		deliveryTime.Day(),
		deliveryTime.Hour(),
		deliveryTime.Minute())
}

// Генерирует случайный набор товаров (от 1 до 5 позиций)
func generateRandomItems() []*pb.OrderItem {
	numItems := 40 // от 1 до 5 товаров
	items := make([]*pb.OrderItem, 0, numItems)
	
	// Чтобы избежать дублирования товаров в одном заказе
	usedProducts := make(map[string]bool)
	
	for len(items) < numItems {
		product := products[rand.Intn(len(products))]
		if !usedProducts[product.ID] {
			quantity := rand.Intn(3) + 1 // от 1 до 3 штук каждого товара
			items = append(items, &pb.OrderItem{
				ProductId: product.ID,
				Name:      product.Name,
				Price:     product.Price,
				Quantity:  int32(quantity),
			})
			usedProducts[product.ID] = true
		}
	}
	
	return items
}

// Вычисляет общую стоимость заказа
func calculateTotalPrice(items []*pb.OrderItem) int32 {
	var total int32
	for _, item := range items {
		total += item.Price * item.Quantity
	}
	return total
}

// Генерирует случайный комментарий курьеру
func generateRandomComment() string {
	comments := []string{
		"Позвонить за 10 минут",
		"Не звонить, напишите в WhatsApp",
		"Оставить у двери",
		"Отдать консьержу",
		"Позвонить в домофон",
		"",
	}
	return comments[rand.Intn(len(comments))]
}